package orm

import (
	"context"
	"html/template"
	"os"

	"github.com/stretchr/testify/suite"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dbfixture"

	"github.com/ilxqx/vef-framework-go/constants"
	"github.com/ilxqx/vef-framework-go/datetime"
	"github.com/ilxqx/vef-framework-go/id"
)

// User represents a user in the system.
type User struct {
	bun.BaseModel `bun:"table:test_user,alias:u"`
	Model

	Name  string `json:"name"     bun:"name,notnull"`
	Email string `json:"email"    bun:"email,notnull,unique"`
	Age   int16  `json:"age"      bun:"age,notnull,default:0"`
	// Bun applies the struct default even when fixtures explicitly set this field to true, so we
	// avoid declaring a default to keep fixture values intact.
	// IsActive bool           `json:"isActive" bun:"is_active,notnull,default:TRUE"`
	IsActive bool           `json:"isActive" bun:"is_active,notnull"`
	Meta     map[string]any `json:"meta"     bun:"meta"`

	// Relations
	Posts []Post `json:"posts" bun:"rel:has-many,join:id=user_id"`
}

// Post represents a blog post or article.
type Post struct {
	bun.BaseModel `bun:"table:test_post,alias:p"`
	Model

	Title       string  `json:"title"       bun:"title,notnull"`
	Content     string  `json:"content"     bun:"content,notnull"`
	Description *string `json:"description" bun:"description"`
	UserId      string  `json:"userId"      bun:"user_id,notnull"`
	CategoryId  string  `json:"categoryId"  bun:"category_id,notnull"`
	Status      string  `json:"status"      bun:"status,notnull,default:'draft'"`
	ViewCount   int     `json:"viewCount"   bun:"view_count,notnull,default:0"`

	// Relations
	User     *User     `json:"user"     bun:"rel:belongs-to,join:user_id=id"`
	Category *Category `json:"category" bun:"rel:belongs-to,join:category_id=id"`
}

// Tag represents a content tag.
type Tag struct {
	bun.BaseModel `bun:"table:test_tag,alias:t"`
	Model

	Name        string  `json:"name"        bun:"name,notnull,unique"`
	Description *string `json:"description" bun:"description"`
}

// PostTag represents the many-to-many relationship between posts and tags.
type PostTag struct {
	bun.BaseModel `bun:"table:test_post_tag,alias:pt"`
	Model

	PostId string `json:"postId" bun:"post_id,notnull"`
	TagId  string `json:"tagId"  bun:"tag_id,notnull"`

	// Relations
	Post *Post `json:"post" bun:"rel:belongs-to,join:post_id=id"`
	Tag  *Tag  `json:"tag"  bun:"rel:belongs-to,join:tag_id=id"`
}

// Category represents a content category.
type Category struct {
	bun.BaseModel `bun:"table:test_category,alias:c"`
	Model

	Name        string  `json:"name"        bun:"name,notnull,unique"`
	Description *string `json:"description" bun:"description"`
	ParentId    *string `json:"parentId"    bun:"parent_id"`

	// Relations
	Posts    []Post     `json:"posts"    bun:"rel:has-many,join:id=category_id"`
	Parent   *Category  `json:"parent"   bun:"rel:belongs-to,join:parent_id=id"`
	Children []Category `json:"children" bun:"rel:has-many,join:id=parent_id"`
}

// SimpleModel represents a simple test model for subquery tests.
type SimpleModel struct {
	bun.BaseModel `bun:"table:test_simple,alias:s"`
	Model

	Name  string `json:"name"  bun:"name,notnull"`
	Value int    `json:"value" bun:"value,notnull"`
}

// OrmTestSuite contains all the actual test methods and works with orm.Db interface.
// This suite will be run against multiple databases to verify cross-database compatibility.
type OrmTestSuite struct {
	suite.Suite

	ctx    context.Context
	db     Db
	dbType constants.DbType
}

// SetupSuite initializes the test suite (called once per database).
func (suite *OrmTestSuite) SetupSuite() {
	suite.T().Logf("Setting up Orm test suite for %s", suite.dbType)

	db := suite.getBunDb()
	db.RegisterModel(
		(*User)(nil),
		(*Post)(nil),
		(*Tag)(nil),
		(*PostTag)(nil),
		(*Category)(nil),
		(*SimpleModel)(nil),
	)

	// Counter for monotonically increasing time offsets
	// This ensures updated_at is always >= created_at since they are called sequentially in fixtures
	counter := 0
	fixture := dbfixture.New(
		db,
		dbfixture.WithRecreateTables(),
		dbfixture.WithTemplateFuncs(template.FuncMap{
			"id": func() string {
				return id.Generate()
			},
			"now": func() string {
				counter++
				// Use monotonically increasing hour offsets to ensure:
				// 1. All timestamps are recent (within a reasonable range from now)
				// 2. Later calls always produce later timestamps (updated_at >= created_at)
				// 3. Meaningful intervals exist for Age function testing
				// Start from -12 hours and increment by 1 hour for each call
				offsetHours := counter - 12

				return datetime.Now().AddHours(offsetHours).String()
			},
		}),
	)

	err := fixture.Load(suite.ctx, os.DirFS("testdata"), "fixture.yaml")
	suite.Require().NoError(err, "Failed to load fixtures")

	_, err = db.NewCreateTable().IfNotExists().Model((*SimpleModel)(nil)).Exec(suite.ctx)
	suite.Require().NoError(err, "Failed to create simple model table")
	suite.Require().NoError(err, "Failed to create complex model table")

	suite.T().Logf("Test fixtures loaded for %s database", suite.dbType)
}

// getBunDb extracts the underlying bun.DB from orm.Db interface.
func (suite *OrmTestSuite) getBunDb() *bun.DB {
	if db, ok := suite.db.(*BunDb); ok {
		if bunDB, ok := db.db.(*bun.DB); ok {
			return bunDB
		}
	}

	suite.Require().Fail("Could not extract bun.DB from orm.Db interface")

	return nil
}

// Helper methods for common test patterns

// AssertCount verifies the count result of a select query.
func (suite *OrmTestSuite) AssertCount(query SelectQuery, expectedCount int64) {
	count, err := query.Count(suite.ctx)
	suite.NoError(err)
	suite.Equal(expectedCount, count, "Count mismatch for %s", suite.dbType)
}

// AssertExists verifies that a query returns at least one result.
func (suite *OrmTestSuite) AssertExists(query SelectQuery) {
	exists, err := query.Exists(suite.ctx)
	suite.NoError(err)
	suite.True(exists, "Query should return results for %s", suite.dbType)
}

// AssertNotExists verifies that a query returns no results.
func (suite *OrmTestSuite) AssertNotExists(query SelectQuery) {
	exists, err := query.Exists(suite.ctx)
	suite.NoError(err)
	suite.False(exists, "Query should not return results for %s", suite.dbType)
}
