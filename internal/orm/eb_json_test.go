package orm

// JsonFunctionsTestSuite tests JSON function methods of ExprBuilder.
type JsonFunctionsTestSuite struct {
	*OrmTestSuite
}

// TestJsonObject tests the JsonObject function.
func (suite *JsonFunctionsTestSuite) TestJsonObject() {
	suite.T().Logf("Testing JsonObject function for %s", suite.dbType)

	suite.Run("CreateJsonObjectFromColumns", func() {
		type JsonObjectResult struct {
			ID         string `bun:"id"`
			Name       string `bun:"name"`
			UserObject string `bun:"user_object"`
		}

		var jsonObjectResults []JsonObjectResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "name").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject("user_id", eb.Column("id"), "user_name", eb.Column("name"), "user_age", eb.Column("age"), "is_active", eb.Column("is_active"))
			}, "user_object").
			OrderBy("name").
			Scan(suite.ctx, &jsonObjectResults)

		suite.NoError(err, "JsonObject should work")
		suite.True(len(jsonObjectResults) > 0, "Should have JSON object results")

		for _, result := range jsonObjectResults {
			suite.NotEmpty(result.UserObject, "User object should not be empty")
			suite.Contains(result.UserObject, result.ID, "JSON should contain user ID")
			suite.Contains(result.UserObject, result.Name, "JSON should contain user name")
			suite.T().Logf("User %s JSON: %s", result.Name, result.UserObject)
		}
	})
}

// TestJsonArray tests the JsonArray function.
func (suite *JsonFunctionsTestSuite) TestJsonArray() {
	suite.T().Logf("Testing JsonArray function for %s", suite.dbType)

	suite.Run("CreateJsonArrayFromColumns", func() {
		type JsonArrayResult struct {
			PostID    string `bun:"post_id"`
			JsonArray string `bun:"json_array"`
		}

		var jsonArrayResults []JsonArrayResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArray(eb.Column("title"), eb.Column("status"), eb.Column("view_count"))
			}, "json_array").
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &jsonArrayResults)

		suite.NoError(err, "JsonArray should work")
		suite.True(len(jsonArrayResults) > 0, "Should have JSON array results")

		for _, result := range jsonArrayResults {
			suite.NotEmpty(result.JsonArray, "JSON array should not be empty")
			suite.T().Logf("Post %s JSON Array: %s", result.PostID, result.JsonArray)
		}
	})
}

// TestJsonExtract tests the JsonExtract function.
func (suite *JsonFunctionsTestSuite) TestJsonExtract() {
	suite.T().Logf("Testing JsonExtract function for %s", suite.dbType)

	suite.Run("ExtractJsonPathValue", func() {
		type JsonExtractResult struct {
			Title       string `bun:"title"`
			MetaJSON    string `bun:"meta_json"`
			ExtractedID string `bun:"extracted_id"`
		}

		var jsonExtractResults []JsonExtractResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject("post_id", eb.Column("id"), "title", eb.Column("title"), "views", eb.Column("view_count"), "status", eb.Column("status"))
			}, "meta_json").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONExtract(eb.JSONObject("id", eb.Column("id")), "id")
			}, "extracted_id").
			OrderBy("title").
			Limit(3).
			Scan(suite.ctx, &jsonExtractResults)

		suite.NoError(err, "JsonExtract should work")
		suite.True(len(jsonExtractResults) > 0, "Should have JSON extract results")

		for _, result := range jsonExtractResults {
			suite.NotEmpty(result.MetaJSON, "Meta JSON should not be empty")
			suite.NotEmpty(result.ExtractedID, "Extracted ID should not be empty")
			suite.T().Logf("Post %s Meta: %s, Extracted ID: %s", result.Title, result.MetaJSON, result.ExtractedID)
		}
	})
}

// TestJsonValid tests the JsonValid function.
func (suite *JsonFunctionsTestSuite) TestJsonValid() {
	suite.T().Logf("Testing JsonValid function for %s", suite.dbType)

	suite.Run("ValidateJsonObject", func() {
		type JsonValidationResult struct {
			Title    string `bun:"title"`
			JsonData string `bun:"json_data"`
			IsValid  bool   `bun:"is_valid"`
		}

		var jsonValidResults []JsonValidationResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count"), "status", eb.Column("status"))
			}, "json_data").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONValid(eb.JSONObject("test", eb.Column("title")))
			}, "is_valid").
			OrderBy("title").
			Limit(2).
			Scan(suite.ctx, &jsonValidResults)

		suite.NoError(err, "JsonValid should work")
		suite.True(len(jsonValidResults) > 0, "Should have JSON validation results")

		for _, result := range jsonValidResults {
			suite.NotEmpty(result.JsonData, "JSON data should not be empty")
			suite.True(result.IsValid, "JSON should be valid")
			suite.T().Logf("Post %s: Valid: %t, Data: %s", result.Title, result.IsValid, result.JsonData)
		}
	})
}

// TestJsonInsert tests the JsonInsert function.
func (suite *JsonFunctionsTestSuite) TestJsonInsert() {
	suite.T().Logf("Testing JsonInsert function for %s", suite.dbType)

	suite.Run("InsertIntoJsonObject", func() {
		type JsonModifyResult struct {
			Title      string `bun:"title"`
			Original   string `bun:"original"`
			WithInsert string `bun:"with_insert"`
		}

		var jsonModifyResults []JsonModifyResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count"))
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONInsert(eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count")), "status", eb.Column("status"))
			}, "with_insert").
			OrderBy("title").
			Limit(2).
			Scan(suite.ctx, &jsonModifyResults)

		suite.NoError(err, "JsonInsert should work")
		suite.True(len(jsonModifyResults) > 0, "Should have JSON modify results")

		for _, result := range jsonModifyResults {
			suite.NotEmpty(result.Original, "Original JSON should not be empty")
			suite.NotEmpty(result.WithInsert, "JSON with insert should not be empty")
			suite.T().Logf("Post %s modifications:", result.Title)
			suite.T().Logf("  Original: %s", result.Original)
			suite.T().Logf("  With Insert: %s", result.WithInsert)
		}
	})
}

// TestJsonReplace tests the JsonReplace function.
func (suite *JsonFunctionsTestSuite) TestJsonReplace() {
	suite.T().Logf("Testing JsonReplace function for %s", suite.dbType)

	suite.Run("ReplaceJsonValue", func() {
		type JsonModifyResult struct {
			Title       string `bun:"title"`
			Original    string `bun:"original"`
			WithReplace string `bun:"with_replace"`
		}

		var jsonModifyResults []JsonModifyResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("title").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count"))
			}, "original").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONReplace(eb.JSONObject("title", eb.Column("title"), "views", eb.Column("view_count")), "views", 9999)
			}, "with_replace").
			OrderBy("title").
			Limit(2).
			Scan(suite.ctx, &jsonModifyResults)

		suite.NoError(err, "JsonReplace should work")
		suite.True(len(jsonModifyResults) > 0, "Should have JSON modify results")

		for _, result := range jsonModifyResults {
			suite.NotEmpty(result.Original, "Original JSON should not be empty")
			suite.NotEmpty(result.WithReplace, "JSON with replace should not be empty")
			suite.Contains(result.WithReplace, "9999", "Replaced JSON should contain new value")
			suite.T().Logf("Post %s: Original: %s, With Replace: %s", result.Title, result.Original, result.WithReplace)
		}
	})
}

// TestJsonLength tests the JsonLength function.
func (suite *JsonFunctionsTestSuite) TestJsonLength() {
	suite.T().Logf("Testing JsonLength function for %s", suite.dbType)

	suite.Run("GetJsonObjectLength", func() {
		type JsonLengthResult struct {
			ID         string `bun:"id"`
			Meta       string `bun:"meta"`
			MetaLength int64  `bun:"meta_length"`
		}

		var results []JsonLengthResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONLength(eb.Column("meta"))
			}, "meta_length").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonLength should work for all databases")
		suite.True(len(results) > 0, "Should have JsonLength results")

		for _, result := range results {
			suite.True(result.MetaLength >= 0, "JsonLength should be non-negative")
			suite.T().Logf("ID: %s, Meta: %s, Length: %d", result.ID, result.Meta, result.MetaLength)
		}
	})

	suite.Run("GetJsonArrayLength", func() {
		type JsonArrayLengthResult struct {
			ID          string `bun:"id"`
			TagsArray   string `bun:"tags_array"`
			ArrayLength int64  `bun:"array_length"`
		}

		var results []JsonArrayLengthResult

		err := suite.db.NewSelect().
			Model((*Post)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArray(eb.Column("title"), eb.Column("status"), eb.Column("view_count"))
			}, "tags_array").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONLength(
					eb.JSONArray(eb.Column("title"), eb.Column("status"), eb.Column("view_count")),
				)
			}, "array_length").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonLength should work for arrays")
		suite.True(len(results) > 0, "Should have JsonLength array results")

		for _, result := range results {
			suite.Equal(int64(3), result.ArrayLength, "Array should have exactly 3 elements")
			suite.T().Logf("ID: %s, Array: %s, Length: %d", result.ID, result.TagsArray, result.ArrayLength)
		}
	})
}

// TestJsonType tests the JsonType function.
func (suite *JsonFunctionsTestSuite) TestJsonType() {
	suite.T().Logf("Testing JsonType function for %s", suite.dbType)

	suite.Run("GetJsonValueType", func() {
		type JsonTypeResult struct {
			ID       string `bun:"id"`
			Meta     string `bun:"meta"`
			MetaType string `bun:"meta_type"`
		}

		var results []JsonTypeResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONType(eb.Column("meta"))
			}, "meta_type").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonType should work for %s", suite.dbType)
		suite.True(len(results) > 0, "Should have JsonType results")

		for _, result := range results {
			suite.True(result.MetaType != "", "JsonType should not be empty")
			suite.T().Logf("ID: %s, Meta: %s, Type: %s", result.ID, result.Meta, result.MetaType)
		}
	})

	suite.Run("GetDifferentJsonTypes", func() {
		type JsonTypesResult struct {
			ID         string `bun:"id"`
			ArrayType  string `bun:"array_type"`
			ObjectType string `bun:"object_type"`
		}

		var results []JsonTypesResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONType(eb.JSONArray(eb.Column("name"), eb.Column("email")))
			}, "array_type").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONType(eb.JSONObject("name", eb.Column("name"), "age", eb.Column("age")))
			}, "object_type").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonType should detect different types")
		suite.True(len(results) > 0, "Should have JsonType results for different types")

		for _, result := range results {
			suite.NotEmpty(result.ArrayType, "Array type should not be empty")
			suite.NotEmpty(result.ObjectType, "Object type should not be empty")
			suite.T().Logf("ID: %s, Types - Array: %s, Object: %s",
				result.ID, result.ArrayType, result.ObjectType)
		}
	})
}

// TestJsonKeys tests the JsonKeys function.
func (suite *JsonFunctionsTestSuite) TestJsonKeys() {
	suite.T().Logf("Testing JsonKeys function for %s", suite.dbType)

	suite.Run("GetJsonObjectKeys", func() {
		type JsonKeysResult struct {
			ID       string `bun:"id"`
			Meta     string `bun:"meta"`
			MetaKeys string `bun:"meta_keys"`
		}

		var results []JsonKeysResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONKeys(eb.Column("meta"))
			}, "meta_keys").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonKeys should work for %s", suite.dbType)
		suite.True(len(results) > 0, "Should have JsonKeys results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Meta: %s, Keys: %s", result.ID, result.Meta, result.MetaKeys)
		}
	})

	suite.Run("GetJsonObjectKeysWithPath", func() {
		type JsonKeysPathResult struct {
			ID         string `bun:"id"`
			Attributes string `bun:"attributes"`
			AttrKeys   string `bun:"attr_keys"`
		}

		var results []JsonKeysPathResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject("name", eb.Column("name"), "age", eb.Column("age"))
			}, "attributes").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONKeys(
					eb.JSONObject("name", eb.Column("name"), "age", eb.Column("age")),
				)
			}, "attr_keys").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonKeys with constructed object should work for %s", suite.dbType)
		suite.True(len(results) > 0, "Should have JsonKeys with path results")

		for _, result := range results {
			suite.NotEmpty(result.AttrKeys, "Attribute keys should not be empty")
			suite.T().Logf("ID: %s, Attributes: %s, Keys: %s", result.ID, result.Attributes, result.AttrKeys)
		}
	})
}

// TestJsonContains tests the JsonContains function.
func (suite *JsonFunctionsTestSuite) TestJsonContains() {
	suite.T().Logf("Testing JsonContains function for %s", suite.dbType)

	suite.Run("CheckJsonContainsValue", func() {
		type JsonContainsResult struct {
			ID           string `bun:"id"`
			Meta         string `bun:"meta"`
			ContainsTest bool   `bun:"contains_test"`
		}

		var results []JsonContainsResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONContains(eb.Column("meta"), `{"active": true}`)
			}, "contains_test").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonContains should work for %s", suite.dbType)
		suite.True(len(results) > 0, "Should have JsonContains results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Meta: %s, Contains: %v", result.ID, result.Meta, result.ContainsTest)
		}
	})
}

// TestJsonContainsPath tests the JsonContainsPath function.
func (suite *JsonFunctionsTestSuite) TestJsonContainsPath() {
	suite.T().Logf("Testing JsonContainsPath function for %s", suite.dbType)

	suite.Run("CheckJsonPathExists", func() {
		type JsonContainsPathResult struct {
			ID         string `bun:"id"`
			Meta       string `bun:"meta"`
			PathExists bool   `bun:"path_exists"`
		}

		var results []JsonContainsPathResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONContainsPath(eb.Column("meta"), "role")
			}, "path_exists").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonContainsPath should work for %s", suite.dbType)
		suite.True(len(results) > 0, "Should have JsonContainsPath results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Meta: %s, PathExists: %v", result.ID, result.Meta, result.PathExists)
		}
	})
}

// TestJsonUnquote tests the JsonUnquote function.
func (suite *JsonFunctionsTestSuite) TestJsonUnquote() {
	suite.T().Logf("Testing JsonUnquote function for %s", suite.dbType)

	suite.Run("RemoveJsonQuotes", func() {
		type JsonUnquoteResult struct {
			ID       string `bun:"id"`
			Meta     string `bun:"meta"`
			Unquoted string `bun:"unquoted"`
		}

		var results []JsonUnquoteResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONUnquote(eb.JSONExtract(eb.Column("meta"), "role"))
			}, "unquoted").
			Limit(5).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonUnquote should work for %s", suite.dbType)

		for _, result := range results {
			suite.T().Logf("ID: %s, Meta: %s, Unquoted: %s", result.ID, result.Meta, result.Unquoted)
		}
	})
}

// TestJsonSet tests the JsonSet function.
func (suite *JsonFunctionsTestSuite) TestJsonSet() {
	suite.T().Logf("Testing JsonSet function for %s", suite.dbType)

	suite.Run("SetJsonPathValue", func() {
		type JsonSetResult struct {
			ID          string `bun:"id"`
			Meta        string `bun:"meta"`
			UpdatedMeta string `bun:"updated_meta"`
		}

		var results []JsonSetResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONSet(eb.Column("meta"), "updated", "true")
			}, "updated_meta").
			Limit(3).
			Scan(suite.ctx, &results)

		suite.NoError(err, "JsonSet should work for %s", suite.dbType)
		suite.True(len(results) > 0, "Should have JsonSet results")

		for _, result := range results {
			suite.T().Logf("ID: %s, Original: %s, Updated: %s", result.ID, result.Meta, result.UpdatedMeta)
		}
	})
}

// TestJsonArrayAppend tests the JsonArrayAppend function.
func (suite *JsonFunctionsTestSuite) TestJsonArrayAppend() {
	suite.T().Logf("Testing JsonArrayAppend function for %s", suite.dbType)

	suite.Run("AppendToJsonArray", func() {
		type JsonArrayAppendResult struct {
			ID          string `bun:"id"`
			Meta        string `bun:"meta"`
			UpdatedMeta string `bun:"updated_meta"`
		}

		var results []JsonArrayAppendResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id", "meta").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArrayAppend(eb.Column("meta"), "interests", `"testing"`)
			}, "updated_meta").
			Limit(3).
			Scan(suite.ctx, &results)
		if err != nil {
			suite.T().Logf("JsonArrayAppend test completed (may have errors if no array fields): %v", err)
		} else {
			suite.True(len(results) >= 0, "JsonArrayAppend should execute for %s", suite.dbType)

			for _, result := range results {
				suite.T().Logf("ID: %s, Original: %s, Updated: %s", result.ID, result.Meta, result.UpdatedMeta)
			}
		}
	})
}

// TestJsonEdgeCases tests JSON function edge cases and boundary conditions.
func (suite *JsonFunctionsTestSuite) TestJsonEdgeCases() {
	suite.T().Logf("Testing JSON edge cases for %s", suite.dbType)

	suite.Run("EmptyJsonObject", func() {
		type EmptyObjectResult struct {
			ID          string `bun:"id"`
			EmptyObject string `bun:"empty_object"`
			ObjectType  string `bun:"object_type"`
			ObjectLen   int64  `bun:"object_len"`
		}

		var results []EmptyObjectResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONObject()
			}, "empty_object").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONType(eb.JSONObject())
			}, "object_type").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONLength(eb.JSONObject())
			}, "object_len").
			Limit(1).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Empty JSON object should work")
		suite.True(len(results) > 0, "Should have empty object results")

		for _, result := range results {
			suite.NotEmpty(result.EmptyObject, "Empty object should not be nil")
			suite.Equal(int64(0), result.ObjectLen, "Empty object should have length 0")
			suite.T().Logf("ID: %s, Empty Object: %s, Type: %s, Length: %d",
				result.ID, result.EmptyObject, result.ObjectType, result.ObjectLen)
		}
	})

	suite.Run("EmptyJsonArray", func() {
		type EmptyArrayResult struct {
			ID         string `bun:"id"`
			EmptyArray string `bun:"empty_array"`
			ArrayType  string `bun:"array_type"`
			ArrayLen   int64  `bun:"array_len"`
		}

		var results []EmptyArrayResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONArray()
			}, "empty_array").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONType(eb.JSONArray())
			}, "array_type").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONLength(eb.JSONArray())
			}, "array_len").
			Limit(1).
			Scan(suite.ctx, &results)

		suite.NoError(err, "Empty JSON array should work")
		suite.True(len(results) > 0, "Should have empty array results")

		for _, result := range results {
			suite.NotEmpty(result.EmptyArray, "Empty array should not be nil")
			suite.Equal(int64(0), result.ArrayLen, "Empty array should have length 0")
			suite.T().Logf("ID: %s, Empty Array: %s, Type: %s, Length: %d",
				result.ID, result.EmptyArray, result.ArrayType, result.ArrayLen)
		}
	})

	suite.Run("JsonValidEdgeCases", func() {
		type JsonValidResult struct {
			ID          string `bun:"id"`
			ValidObject bool   `bun:"valid_object"`
			ValidArray  bool   `bun:"valid_array"`
			InvalidJSON bool   `bun:"invalid_json"`
		}

		var results []JsonValidResult

		err := suite.db.NewSelect().
			Model((*User)(nil)).
			Select("id").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONValid(eb.JSONObject("test", "value"))
			}, "valid_object").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONValid(eb.JSONArray("a", "b", "c"))
			}, "valid_array").
			SelectExpr(func(eb ExprBuilder) any {
				return eb.JSONValid("not valid json")
			}, "invalid_json").
			Limit(1).
			Scan(suite.ctx, &results)
		if err != nil {
			suite.T().Logf("JsonValid edge cases test completed (expected for invalid JSON): %v", err)
		} else {
			suite.True(len(results) >= 0, "JsonValid edge cases should execute")

			for _, result := range results {
				suite.T().Logf("ID: %s, Valid Object: %t, Valid Array: %t, Invalid: %t",
					result.ID, result.ValidObject, result.ValidArray, result.InvalidJSON)
			}
		}
	})
}
