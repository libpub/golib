package rdbms

import "xorm.io/xorm/names"

// RelationQuery for relation includes query
// examples: tables: user(id) - user_role_map(user_id, role_id) - role(id)
//
//	RelationTable is user_role_map
//	TargetTable is role
//	SelfRelationField is user_id
//	TargetRelationField is role_id
//	TargetPrimaryKey is (role.)id
type RelationQuery struct {
	Select              string
	RelationTable       names.TableName
	TargetTable         names.TableName
	SelfRelationField   string
	TargetRelationField string
	TargetPrimaryKey    string
}

// DynamicModel for create talbe struct in runtime
type DynamicModel struct {
	customizedTableName string `xorm:"-"`
}

// TableName returns table name
func (m *DynamicModel) TableName() string {
	return m.customizedTableName
}

// SetTableName set table name for table struct created in runtime
func (m *DynamicModel) SetTableName(tableName string) {
	m.customizedTableName = tableName
}

// DynamicModelDefination for dynamic table structure
type DynamicModelDefination struct {
	TableName string             `json:"tableName"`
	Columns   []ColumnDefination `json:"columns"`
}

// ColumnDefination for dynamic table column structure
type ColumnDefination struct {
	Name          string      `json:"name"`
	Type          string      `json:"type"`
	Comment       string      `json:"comment"`
	Index         bool        `json:"index"`
	Unique        bool        `json:"unique"`
	PrimaryKey    bool        `json:"primaryKey"`
	AutoIncrement bool        `json:"autoIncrement"`
	NotNull       bool        `json:"notNull"`
	Default       interface{} `json:"default"`
}
