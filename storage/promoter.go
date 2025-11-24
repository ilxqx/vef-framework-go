package storage

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/event"
	"github.com/ilxqx/vef-framework-go/null"
	"github.com/ilxqx/vef-framework-go/reflectx"
	"github.com/ilxqx/vef-framework-go/set"
	"github.com/ilxqx/vef-framework-go/strhelpers"
)

// MetaType defines the type of meta information field.
type MetaType string

const (
	// MetaTypeUploadedFile indicates a direct file field (string or []string).
	MetaTypeUploadedFile MetaType = "uploaded_file"
	// MetaTypeRichText indicates a rich text field containing HTML with resource references.
	MetaTypeRichText MetaType = "richtext"
	// MetaTypeMarkdown indicates a Markdown field containing resource references.
	MetaTypeMarkdown MetaType = "markdown"
)

const (
	tagMeta = "meta"
)

var nullStringType = reflect.TypeFor[null.String]()

func getStringValue(fieldValue reflect.Value) (string, bool) {
	fieldType := fieldValue.Type()

	if fieldType.Kind() == reflect.String {
		return fieldValue.String(), true
	}

	if fieldType.Kind() == reflect.Pointer &&
		fieldType.Elem().Kind() == reflect.String {

		if fieldValue.IsNil() {
			return constants.Empty, false
		}

		return fieldValue.Elem().String(), true
	}

	if fieldType == nullStringType {
		if ns := fieldValue.Interface().(null.String); ns.Valid {
			return ns.String, true
		}

		return constants.Empty, false
	}

	return constants.Empty, false
}

func setStringValue(fieldValue reflect.Value, value string) {
	fieldType := fieldValue.Type()

	if fieldType.Kind() == reflect.String {
		fieldValue.SetString(value)

		return
	}

	if fieldType.Kind() == reflect.Pointer &&
		fieldType.Elem().Kind() == reflect.String {

		strValue := value
		fieldValue.Set(reflect.ValueOf(&strValue))

		return
	}

	if fieldType == nullStringType {
		fieldValue.Set(reflect.ValueOf(null.StringFrom(value)))
	}
}

func getStringSliceValue(fieldValue reflect.Value) ([]string, bool) {
	fieldType := fieldValue.Type()

	if fieldType.Kind() == reflect.Slice &&
		fieldType.Elem().Kind() == reflect.String {

		if fieldValue.IsNil() {
			return nil, false
		}

		return fieldValue.Interface().([]string), true
	}

	return nil, false
}

func setStringSliceValue(fieldValue reflect.Value, value []string) {
	fieldType := fieldValue.Type()

	if fieldType.Kind() == reflect.Slice &&
		fieldType.Elem().Kind() == reflect.String {

		fieldValue.Set(reflect.ValueOf(value))
	}
}

// metaField represents the configuration of a meta information field.
type metaField struct {
	// Field index in the struct
	index []int
	// Meta type
	typ MetaType
	// Whether it's a []string (only valid for uploaded_file)
	isArray bool
	// Parsed attributes from the meta tag
	attrs map[string]string
}

func isStringType(fieldType reflect.Type) bool {
	if fieldType.Kind() == reflect.String {
		return true
	}

	if fieldType.Kind() == reflect.Pointer &&
		fieldType.Elem().Kind() == reflect.String {

		return true
	}

	return fieldType == nullStringType
}

func isStringSliceType(fieldType reflect.Type) bool {
	return fieldType.Kind() == reflect.Slice &&
		fieldType.Elem().Kind() == reflect.String
}

type defaultPromoter[T any] struct {
	service   Service
	publisher event.Publisher
	fields    []metaField
}

// NewPromoter creates a new Promoter for type T.
// The publisher parameter is optional; if omitted, no events will be published.
func NewPromoter[T any](service Service, publisher ...event.Publisher) Promoter[T] {
	typ := reflectx.Indirect(reflect.TypeFor[T]())

	var pub event.Publisher
	if len(publisher) > 0 {
		pub = publisher[0]
	}

	return &defaultPromoter[T]{
		service:   service,
		publisher: pub,
		fields:    parseMetaFields(typ),
	}
}

func parseMetaFields(typ reflect.Type) []metaField {
	if typ.Kind() != reflect.Struct {
		return nil
	}

	fields := make([]metaField, 0)

	visitor := reflectx.TypeVisitor{
		VisitFieldType: func(field reflect.StructField, depth int) reflectx.VisitAction {
			tag, hasTag := field.Tag.Lookup(tagMeta)
			if !hasTag {
				return reflectx.SkipChildren
			}

			var (
				parsed = strhelpers.ParseTag(tag, strhelpers.WithBareValueMode(strhelpers.BareAsKey))

				metaType      MetaType
				metaTypeValue string
				foundMetaType bool
			)

			for key, value := range parsed {
				if foundMetaType {
					break
				}

				switch MetaType(key) {
				case MetaTypeUploadedFile, MetaTypeRichText, MetaTypeMarkdown:
					metaType = MetaType(key)
					metaTypeValue = value
					foundMetaType = true
				}
			}

			if !foundMetaType {
				return reflectx.SkipChildren
			}

			// Format: "category:gallery public:true" -> {"category": "gallery", "public": "true"}
			attrs := strhelpers.ParseTag(
				metaTypeValue,
				strhelpers.WithSpacePairDelimiter(),
				strhelpers.WithValueDelimiter(constants.ByteColon),
			)

			var (
				fieldType = field.Type
				isArray   bool
			)

			// For "uploaded_file", support both scalar (string/ptr/null) and array ([]string) types.
			// For "richtext" and "markdown", only scalar types are allowed.
			if metaType == MetaTypeUploadedFile {
				if isStringSliceType(fieldType) {
					isArray = true
				} else if !isStringType(fieldType) {
					return reflectx.SkipChildren
				}
			} else {
				if !isStringType(fieldType) {
					return reflectx.SkipChildren
				}
			}

			fields = append(fields, metaField{
				index:   field.Index,
				typ:     metaType,
				isArray: isArray,
				attrs:   attrs,
			})

			return reflectx.SkipChildren
		},
	}

	reflectx.VisitType(
		typ,
		visitor,
		reflectx.WithDiveTag(tagMeta, "dive"),
		reflectx.WithTraversalMode(reflectx.DepthFirst),
	)

	return fields
}

func convertToPermanentKey(key string) string {
	return strings.TrimPrefix(key, TempPrefix)
}

func (p *defaultPromoter[T]) publishEvent(evt event.Event) {
	if p.publisher != nil {
		p.publisher.Publish(evt)
	}
}

func (p *defaultPromoter[T]) Promote(ctx context.Context, newModel, oldModel *T) error {
	switch {
	case newModel != nil && oldModel != nil:
		if err := p.promoteFiles(ctx, newModel); err != nil {
			return err
		}

		return p.cleanupReplacedFiles(ctx, newModel, oldModel)

	case newModel != nil:
		return p.promoteFiles(ctx, newModel)

	case oldModel != nil:
		return p.deleteAllFiles(ctx, oldModel)

	default:
		return nil
	}
}

func (p *defaultPromoter[T]) promoteFiles(ctx context.Context, model *T) error {
	value := reflect.Indirect(reflect.ValueOf(model))
	if value.Kind() != reflect.Struct {
		return nil
	}

	for _, field := range p.fields {
		fieldValue := value.FieldByIndex(field.index)
		if !fieldValue.CanSet() {
			continue
		}

		switch field.typ {
		case MetaTypeUploadedFile:
			if err := p.promoteUploadedFileField(ctx, fieldValue, field.isArray, field.typ, field.attrs); err != nil {
				return err
			}

		case MetaTypeRichText:
			if err := p.promoteContentField(ctx, fieldValue, extractHtmlUrls, replaceHtmlUrls, field.typ, field.attrs); err != nil {
				return err
			}

		case MetaTypeMarkdown:
			if err := p.promoteContentField(ctx, fieldValue, extractMarkdownUrls, replaceMarkdownUrls, field.typ, field.attrs); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *defaultPromoter[T]) promoteUploadedFileField(ctx context.Context, fieldValue reflect.Value, isArray bool, metaType MetaType, attrs map[string]string) error {
	if isArray {
		keys, valid := getStringSliceValue(fieldValue)
		if !valid || len(keys) == 0 {
			return nil
		}

		promotedKeys := make([]string, 0, len(keys))

		for _, key := range keys {
			if key = strings.TrimSpace(key); key == constants.Empty {
				continue
			}

			promotedKey, err := p.promoteSingleFile(ctx, key, metaType, attrs)
			if err != nil {
				return err
			}

			promotedKeys = append(promotedKeys, promotedKey)
		}

		setStringSliceValue(fieldValue, promotedKeys)
	} else {
		key, valid := getStringValue(fieldValue)
		if !valid || key == constants.Empty {
			return nil
		}

		promotedKey, err := p.promoteSingleFile(ctx, key, metaType, attrs)
		if err != nil {
			return err
		}

		setStringValue(fieldValue, promotedKey)
	}

	return nil
}

func (p *defaultPromoter[T]) promoteContentField(
	ctx context.Context,
	fieldValue reflect.Value,
	extractFunc func(string) []string,
	replaceFunc func(string, map[string]string) string,
	metaType MetaType,
	attrs map[string]string,
) error {
	content, valid := getStringValue(fieldValue)
	if !valid || content == constants.Empty {
		return nil
	}

	urls := extractFunc(content)
	if len(urls) == 0 {
		return nil
	}

	// Only promote temp files; permanent files remain unchanged.
	replacements := make(map[string]string)
	for _, url := range urls {
		if !strings.HasPrefix(url, TempPrefix) {
			continue
		}

		promotedKey, err := p.promoteSingleFile(ctx, url, metaType, attrs)
		if err != nil {
			return err
		}

		if promotedKey != url {
			replacements[url] = promotedKey
		}
	}

	if len(replacements) > 0 {
		newContent := replaceFunc(content, replacements)
		setStringValue(fieldValue, newContent)
	}

	return nil
}

func (p *defaultPromoter[T]) promoteSingleFile(ctx context.Context, key string, metaType MetaType, attrs map[string]string) (string, error) {
	if !strings.HasPrefix(key, TempPrefix) {
		return key, nil
	}

	info, err := p.service.PromoteObject(ctx, key)
	if err != nil {
		if errors.Is(err, ErrObjectNotFound) {
			permanentKey := convertToPermanentKey(key)
			if _, err := p.service.StatObject(ctx, StatObjectOptions{Key: permanentKey}); err == nil {
				return permanentKey, nil
			}
		}

		return constants.Empty, fmt.Errorf("failed to promote file %q: %w", key, err)
	}

	if info == nil {
		return key, nil
	}

	p.publishEvent(NewFilePromotedEvent(metaType, info.Key, attrs))

	return info.Key, nil
}

func (p *defaultPromoter[T]) cleanupReplacedFiles(ctx context.Context, newModel, oldModel *T) error {
	oldFiles := p.extractAllFileKeysWithInfo(oldModel)
	newKeys := p.extractAllFileKeys(newModel)

	newSet := set.NewHashSetFromSlice(newKeys)

	for _, fileInfo := range oldFiles {
		if !newSet.Contains(fileInfo.key) {
			if err := p.service.DeleteObject(ctx, DeleteObjectOptions{Key: fileInfo.key}); err != nil {
				return fmt.Errorf("failed to delete file %q: %w", fileInfo.key, err)
			}

			p.publishEvent(NewFileDeletedEvent(fileInfo.metaType, fileInfo.key, fileInfo.attrs))
		}
	}

	return nil
}

func (p *defaultPromoter[T]) deleteAllFiles(ctx context.Context, model *T) error {
	files := p.extractAllFileKeysWithInfo(model)

	for _, fileInfo := range files {
		if fileInfo.key = strings.TrimSpace(fileInfo.key); fileInfo.key == constants.Empty {
			continue
		}

		if err := p.service.DeleteObject(ctx, DeleteObjectOptions{Key: fileInfo.key}); err != nil {
			return fmt.Errorf("failed to delete file %q: %w", fileInfo.key, err)
		}

		p.publishEvent(NewFileDeletedEvent(fileInfo.metaType, fileInfo.key, fileInfo.attrs))
	}

	return nil
}

type fileInfo struct {
	key      string
	metaType MetaType
	attrs    map[string]string
}

func (p *defaultPromoter[T]) extractAllFileKeysWithInfo(model *T) []fileInfo {
	if model == nil {
		return nil
	}

	value := reflect.Indirect(reflect.ValueOf(model))
	if value.Kind() != reflect.Struct {
		return nil
	}

	allFiles := make([]fileInfo, 0)

	for _, field := range p.fields {
		fieldValue := value.FieldByIndex(field.index)
		if !fieldValue.IsValid() {
			continue
		}

		switch field.typ {
		case MetaTypeUploadedFile:
			if field.isArray {
				if keys, valid := getStringSliceValue(fieldValue); valid {
					for _, key := range keys {
						allFiles = append(allFiles, fileInfo{
							key:      key,
							metaType: field.typ,
							attrs:    field.attrs,
						})
					}
				}
			} else {
				if key, valid := getStringValue(fieldValue); valid && key != constants.Empty {
					allFiles = append(allFiles, fileInfo{
						key:      key,
						metaType: field.typ,
						attrs:    field.attrs,
					})
				}
			}

		case MetaTypeRichText:
			if content, valid := getStringValue(fieldValue); valid && content != constants.Empty {
				urls := extractHtmlUrls(content)
				for _, url := range urls {
					allFiles = append(allFiles, fileInfo{
						key:      url,
						metaType: field.typ,
						attrs:    field.attrs,
					})
				}
			}

		case MetaTypeMarkdown:
			if content, valid := getStringValue(fieldValue); valid && content != constants.Empty {
				urls := extractMarkdownUrls(content)
				for _, url := range urls {
					allFiles = append(allFiles, fileInfo{
						key:      url,
						metaType: field.typ,
						attrs:    field.attrs,
					})
				}
			}
		}
	}

	return allFiles
}

func (p *defaultPromoter[T]) extractAllFileKeys(model *T) []string {
	if model == nil {
		return nil
	}

	value := reflect.Indirect(reflect.ValueOf(model))
	if value.Kind() != reflect.Struct {
		return nil
	}

	allKeys := make([]string, 0)

	for _, field := range p.fields {
		fieldValue := value.FieldByIndex(field.index)
		if !fieldValue.IsValid() {
			continue
		}

		switch field.typ {
		case MetaTypeUploadedFile:
			if field.isArray {
				if keys, valid := getStringSliceValue(fieldValue); valid {
					allKeys = append(allKeys, keys...)
				}
			} else {
				if key, valid := getStringValue(fieldValue); valid && key != constants.Empty {
					allKeys = append(allKeys, key)
				}
			}

		case MetaTypeRichText:
			if content, valid := getStringValue(fieldValue); valid && content != constants.Empty {
				urls := extractHtmlUrls(content)
				allKeys = append(allKeys, urls...)
			}

		case MetaTypeMarkdown:
			if content, valid := getStringValue(fieldValue); valid && content != constants.Empty {
				urls := extractMarkdownUrls(content)
				allKeys = append(allKeys, urls...)
			}
		}
	}

	return allKeys
}
