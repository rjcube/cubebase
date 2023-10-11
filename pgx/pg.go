package pgx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/rjcube/cubebase"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Insert 根据数据库结构体进行新增操作
// 本方法依赖于gin线程中的cube对象和SysUserPO对象
func Insert(c *gin.Context, po interface{}) (int64, error) {
	userInfo, _ := c.MustGet("userInfo").(*cubebase.UserInfo)
	userId := fmt.Sprintf("%d", userInfo.ID)
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	sn, pfs := getStructDbFields(po)

	m, err := handleInsertOrUpdateDate(po)
	if nil != err {
		return 0, err
	}

	var buffer bytes.Buffer
	buffer.WriteString("INSERT INTO ")
	buffer.WriteString(getTableName(sn))
	buffer.WriteString(" (")
	flag := false
	for _, pf := range pfs {
		df := pf.C
		if df == "id" || df == "gmt_delete" || df == "delete_user_id" {
			continue
		}
		if !flag {
			flag = true
		} else {
			buffer.WriteString(", ")
		}
		buffer.WriteString(df)
	}
	buffer.WriteString(") VALUES (")
	flag = false
	for _, pf := range pfs {
		df := pf.C
		if df == "id" || df == "gmt_delete" || df == "delete_user_id" {
			continue
		}
		if !flag {
			flag = true
		} else {
			buffer.WriteString(", ")
		}
		buffer.WriteString(getSqlValueField(pf, userId))
	}
	buffer.WriteString(") RETURNING id")

	rows, err := cube.Db.NamedQuery(buffer.String(), m)
	if err != nil {
		fmt.Println(err.Error())
		return 0, &CubePgExecError{Msg: "保存数据失败"}
	}

	id := getInsertID(rows)

	return id, nil
}

// UpdateById 使用记录主键更新PO实体
// 本方法依赖于gin线程中的cube对象和SysUserPO对象
func UpdateById(c *gin.Context, po interface{}) error {
	userInfo, _ := c.MustGet("userInfo").(*cubebase.UserInfo)
	userId := fmt.Sprintf("%d", userInfo.ID)
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	m, err := handleInsertOrUpdateDate(po)
	if nil != err {
		return err
	}

	var buffer bytes.Buffer
	montageUpdate(&buffer, userId, po)
	buffer.WriteString(" AND id = :id")

	_, err = cube.Db.NamedExec(buffer.String(), m)
	if err != nil {
		fmt.Println(err.Error())
		return &CubePgExecError{Msg: "修改数据失败"}
	}

	return nil
}

// UpdateByCondition 根据条件更新数据
// 本方法依赖于gin线程中的cube对象和SysUserPO对象
func UpdateByCondition(c *gin.Context, bp BizParam, po interface{}) error {

	condition, condVal := bp.getCondition(And)
	if condition == "" {
		return &CubePgExecError{Msg: "需要至少一个Where参数"}
	}

	userInfo, _ := c.MustGet("userInfo").(*cubebase.UserInfo)
	userId := fmt.Sprintf("%d", userInfo.ID)
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	var buffer bytes.Buffer
	montageUpdate(&buffer, userId, po)
	buffer.WriteString(" and ")
	buffer.WriteString(condition)

	_, err := cube.Db.NamedExec(buffer.String(), condVal)
	if err != nil {
		fmt.Println(err.Error())
		return &CubePgExecError{Msg: "修改数据失败"}
	}

	return nil
}

// DeleteById 根据主键删除数据
// 本方法依赖于gin线程中的cube对象和SysUserPO对象
func DeleteById(c *gin.Context, po interface{}) error {

	userInfo, _ := c.MustGet("userInfo").(*cubebase.UserInfo)
	userId := fmt.Sprintf("%d", userInfo.ID)
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	sn, _ := getStructDbFields(po)

	var buffer bytes.Buffer
	buffer.WriteString("UPDATE ")
	buffer.WriteString(getTableName(sn))
	buffer.WriteString(" SET gmt_delete = NOW(), is_delete = true, delete_user_id = ")
	buffer.WriteString(userId)
	buffer.WriteString(" WHERE is_delete = false AND id = ")
	buffer.WriteString(":id")

	_, err := cube.Db.NamedExec(buffer.String(), po)
	if err != nil {
		fmt.Println(err.Error())
		return &CubePgExecError{Msg: "修改数据失败"}
	}

	return nil
}

// DeleteByCondition 根据条件删除数据
// 本方法依赖于gin线程中的cube对象和SysUserPO对象
func DeleteByCondition(c *gin.Context, bp BizParam, po interface{}) error {

	userInfo, _ := c.MustGet("userInfo").(*cubebase.UserInfo)
	userId := fmt.Sprintf("%d", userInfo.ID)
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	sn, _ := getStructDbFields(po)

	condition, condVal := bp.getCondition(And)
	if condition == "" {
		return &CubePgExecError{Msg: "需要至少一个Where参数"}
	}

	var buffer bytes.Buffer
	buffer.WriteString("update ")
	buffer.WriteString(getTableName(sn))
	buffer.WriteString(" set is_delete = true, gmt_delete = NOW(), delete_user_id = ")
	buffer.WriteString(userId)
	buffer.WriteString(" where is_delete = false and ")
	buffer.WriteString(condition)

	_, err := cube.Db.NamedExec(buffer.String(), condVal)
	if err != nil {
		return &CubePgExecError{Msg: "删除数据失败"}
	}

	return nil
}

func CountAll[T interface{}](c *gin.Context, bp BizParam, po T) (int64, error) {
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)
	sn, _ := getStructDbFields(po)

	var buffer bytes.Buffer

	buffer.WriteString("select count(0) from ")
	buffer.WriteString(getTableName(sn))
	buffer.WriteString(" where is_delete = false")

	condition, condVal := bp.getCondition(And)
	if condition != "" {
		buffer.WriteString(" and ")
		buffer.WriteString(condition)
	}

	fmt.Println("待执行的SQL语句内容为：" + buffer.String())

	rows, err := cube.Db.NamedQuery(buffer.String(), condVal)
	if nil != err {
		errMsg := err.Error()
		fmt.Println(errMsg)
		if "sql: no rows in result set" == errMsg {
			return 0, nil
		}
		return -1, &CubePgExecError{Msg: "查询统计数据失败"}
	}
	result := countResult{Count: -1}
	for rows.Next() {
		if err := rows.StructScan(&result); err != nil {

		}
		break
	}
	if result.Count == -1 {
		return -1, &CubePgExecError{Msg: "查询统计数据失败"}
	}

	return result.Count, nil
}

// QueryAll 根据条件查询数据
// 本方法依赖于gin线程中的cube对象
func QueryAll[T interface{}](c *gin.Context, bp BizParam, po T) ([]T, error) {

	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	var buffer bytes.Buffer

	montageQuery(&buffer, po)

	condition, condVal := bp.getCondition(And)
	if condition != "" {
		buffer.WriteString(" and ")
		buffer.WriteString(condition)
	}

	buffer.WriteString(" ")
	buffer.WriteString(bp.getOrderBy())

	buffer.WriteString(" ")
	buffer.WriteString(bp.getPage())

	fmt.Println("查询数据SQL为：" + buffer.String())

	rows, err := cube.Db.NamedQuery(buffer.String(), condVal)
	if err != nil {
		fmt.Println(err.Error())
		return nil, &CubePgExecError{Msg: "查询数据失败"}
	}

	var ts []T
	for rows.Next() {
		var t T
		if err := rows.StructScan(&t); err != nil {

		}
		ts = append(ts, t)
	}

	return ts, nil
}

// QueryOne 根据条件查询一条数据，如果数据为多条返回错误信息
// 本方法依赖于gin线程中的cube对象
func QueryOne[T interface{}](c *gin.Context, bp BizParam, po T) (*T, error) {
	ts, err := QueryAll(c, bp, po)
	if nil != err {
		return nil, err
	}
	if nil == ts || len(ts) == 0 {
		return nil, nil
	}
	if len(ts) > 1 {
		return nil, &CubePgExecError{Msg: "查询数据大于一条"}
	}
	return &ts[0], nil
}

func QueryById[T interface{}](c *gin.Context, po T) (*T, error) {
	cube, _ := c.MustGet("cube").(*cubebase.CubeContext)

	var buffer bytes.Buffer

	montageQuery(&buffer, po)

	buffer.WriteString(" and id = :id")

	rows, err := cube.Db.NamedQuery(buffer.String(), po)
	if err != nil {
		fmt.Println(err.Error())
		return nil, &CubePgExecError{Msg: "修改数据失败"}
	}

	var t T
	for rows.Next() {
		if err := rows.StructScan(&t); err != nil {

		}
	}

	return &t, nil
}

func handleInsertOrUpdateDate(po interface{}) (map[string]interface{}, error) {
	poBytes, _ := json.Marshal(po)
	var m map[string]interface{}
	err := json.Unmarshal(poBytes, &m)
	if err != nil {
		return m, &CubePgExecError{Msg: "处理保存数据失败"}
	}

	for k, v := range m {
		vt := reflect.TypeOf(v)
		tName := vt.Name()
		if tName == "Int64Array" || tName == "*Int64Array" || tName == "StringArray" || tName == "*StringArray" {
			m[k] = pq.Array(v)
		}
	}
	return m, nil
}

type BizParam struct {
	condition        []paramCondition
	ands             []BizParam
	ors              []BizParam
	pageLimit        *cubebase.PageForm
	orderBys         []orderBy
	nilConditionSkip bool
}

func NewBizParamByNilSkip(nilConditionSkip bool) BizParam {
	bp := BizParam{
		condition:        nil,
		ands:             nil,
		ors:              nil,
		pageLimit:        nil,
		orderBys:         nil,
		nilConditionSkip: nilConditionSkip,
	}
	return bp
}

func NewBizParam() BizParam {
	return NewBizParamByNilSkip(true)
}

func (bp BizParam) Equal(fn string, val interface{}) BizParam {
	if bp.nilConditionSkip && nil == val {
		return bp
	}
	return bp.addCond(fn, equal, val)
}

func (bp BizParam) NotEqual(fn string, val interface{}) BizParam {
	if bp.nilConditionSkip && nil == val {
		return bp
	}
	return bp.addCond(fn, notEqual, val)
}

func (bp BizParam) GreaterThan(fn string, val interface{}) BizParam {
	if bp.nilConditionSkip && nil == val {
		return bp
	}
	return bp.addCond(fn, greaterThan, val)
}

func (bp BizParam) GreaterThanOrEqual(fn string, val interface{}) BizParam {
	if bp.nilConditionSkip && nil == val {
		return bp
	}
	return bp.addCond(fn, greaterThanOrEqual, val)
}

func (bp BizParam) LessThan(fn string, val interface{}) BizParam {
	if bp.nilConditionSkip && nil == val {
		return bp
	}
	return bp.addCond(fn, lessThan, val)
}

func (bp BizParam) LessThanOrEqual(fn string, val interface{}) BizParam {
	if bp.nilConditionSkip && nil == val {
		return bp
	}
	return bp.addCond(fn, lessThanOrEqual, val)
}

func (bp BizParam) In(fn string, val []interface{}) BizParam {
	if bp.nilConditionSkip && (nil == val || 0 == len(val)) {
		return bp
	}
	return bp.addCond(fn, in, pq.Array(val))
}

func (bp BizParam) NotIn(fn string, val []interface{}) BizParam {
	if bp.nilConditionSkip && (nil == val || 0 == len(val)) {
		return bp
	}
	return bp.addCond(fn, notIn, pq.Array(val))
}

func (bp BizParam) Like(fn string, val *string) BizParam {
	if bp.nilConditionSkip && (nil == val || "" == *val) {
		return bp
	}
	return bp.addCond(fn, like, val)
}

func (bp BizParam) LikeLeft(fn string, val *string) BizParam {
	if bp.nilConditionSkip && (nil == val || "" == *val) {
		return bp
	}
	return bp.addCond(fn, likeLeft, val)
}

func (bp BizParam) LikeRight(fn string, val *string) BizParam {
	if bp.nilConditionSkip && (nil == val || "" == *val) {
		return bp
	}
	return bp.addCond(fn, likeRight, val)
}

func (bp BizParam) NotLike(fn string, val *string) BizParam {
	if bp.nilConditionSkip && (nil == val || "" == *val) {
		return bp
	}
	return bp.addCond(fn, notLike, val)
}

func (bp BizParam) NotLikeLeft(fn string, val *string) BizParam {
	if bp.nilConditionSkip && (nil == val || "" == *val) {
		return bp
	}
	return bp.addCond(fn, notLikeLeft, val)
}

func (bp BizParam) NotLikeRight(fn string, val *string) BizParam {
	if bp.nilConditionSkip && (nil == val || "" == *val) {
		return bp
	}
	return bp.addCond(fn, notLikeRight, val)
}

func (bp BizParam) IsNull(fn string) BizParam {
	return bp.addCond(fn, isNull, nil)
}

func (bp BizParam) IsNotNull(fn string) BizParam {
	return bp.addCond(fn, isNotNull, nil)
}

func (bp BizParam) IsEmpty(fn string) BizParam {
	return bp.addCond(fn, isEmpty, nil)
}

func (bp BizParam) IsNotEmpty(fn string) BizParam {
	return bp.addCond(fn, isNotEmpty, nil)
}

func (bp BizParam) Contain(fn string, val []interface{}) BizParam {
	if bp.nilConditionSkip && (nil == val || 0 == len(val)) {
		return bp
	}
	return bp.addCond(fn, contain, pq.Array(val))
}

func (bp BizParam) ContainAny(fn string, val []interface{}) BizParam {
	if bp.nilConditionSkip && (nil == val || 0 == len(val)) {
		return bp
	}
	return bp.addCond(fn, containAny, pq.Array(val))
}

func (bp BizParam) NotContain(fn string, val []interface{}) BizParam {
	if bp.nilConditionSkip && (nil == val || 0 == len(val)) {
		return bp
	}
	return bp.addCond(fn, notContain, pq.Array(val))
}

func (bp BizParam) BetweenAnd(fn string, from interface{}, to interface{}) BizParam {
	if bp.nilConditionSkip && (nil == from || nil == to) {
		return bp
	}
	val := between{
		From: from,
		To:   to,
	}
	return bp.addCond(fn, betweenAnd, val)
}

func (bp BizParam) And(and BizParam) BizParam {
	bp.ands = append(bp.ands, and)
	return bp
}

func (bp BizParam) Or(and BizParam) BizParam {
	bp.ors = append(bp.ors, and)
	return bp
}

func (bp BizParam) Page(pageIndex int64, pageSize int64) BizParam {
	bp.pageLimit = &cubebase.PageForm{
		PageIndex: pageIndex,
		PageSize:  pageSize,
	}
	return bp
}

func (bp BizParam) OrderBy(fn string, sort Sort) BizParam {
	orderBy := orderBy{
		Column: fn,
		Sort:   sort,
	}
	bp.orderBys = append(bp.orderBys, orderBy)
	return bp
}

func montageUpdate(buffer *bytes.Buffer, userId string, po interface{}) {
	sn, pfs := getStructDbFields(po)

	buffer.WriteString("UPDATE ")
	buffer.WriteString(getTableName(sn))
	buffer.WriteString(" SET ")
	flag := false
	for _, pf := range pfs {
		if pf.C == "gmt_create" || pf.C == "create_user_id" || pf.C == "id" ||
			pf.C == "is_delete" || pf.C == "gmt_delete" {
			continue
		}
		if !flag {
			flag = true
		} else {
			buffer.WriteString(", ")
		}
		buffer.WriteString(pf.C)
		buffer.WriteString(" = ")
		buffer.WriteString(getSqlValueField(pf, userId))
	}
	buffer.WriteString(" WHERE is_delete = false")
}

func montageQuery(buffer *bytes.Buffer, po interface{}) {
	sn, pfs := getStructDbFields(po)

	buffer.WriteString("select ")

	flag := false
	for _, pf := range pfs {
		if !flag {
			flag = true
		} else {
			buffer.WriteString(", ")
		}
		buffer.WriteString(pf.C)
	}

	buffer.WriteString(" from ")
	buffer.WriteString(getTableName(sn))
	buffer.WriteString(" where is_delete = false")
}

func getTableName(sn string) string {
	tableName := camelToField(sn)
	return tableName[0:strings.Index(tableName, "_")] + "." + tableName
}

func getStructDbFields(po interface{}) (string, []pgField) {
	pt := reflect.TypeOf(po)
	if pt.Kind() == reflect.Ptr {
		pt = pt.Elem()
	}
	var dbFields []pgField
	for i := 0; i < pt.NumField(); i++ {
		f := pt.Field(i)
		ft := f.Type.String()
		db := f.Tag.Get("db")
		if "" == db {
			continue
		}
		pf := pgField{
			C: db,
			T: ft,
		}
		dbFields = append(dbFields, pf)
	}
	return pt.Name(), dbFields
}

func getSqlValueField(pf pgField, userId string) string {
	if pf.C == "gmt_create" || pf.C == "gmt_modified" {
		return "NOW()"
	} else if pf.C == "is_delete" {
		return "false"
	} else if pf.C == "create_user_id" || pf.C == "modify_user_id" {
		return userId
	} else if pf.T == "json.RawMessage" || pf.T == "*json.RawMessage" {
		return "CAST(:" + pf.C + " AS jsonb)"
	} else {
		return ":" + pf.C
	}
}

type pgField struct {
	C string
	T string
}

type paramCondition struct {
	Colum  string
	Symbol Symbol
	Value  interface{}
}

func (bp BizParam) getCondition(cs ConnSymbol) (string, map[string]interface{}) {
	var buffer bytes.Buffer
	buffer.WriteString("")
	condVal := make(map[string]interface{})
	conditions := bp.condition
	if nil != conditions && len(conditions) != 0 {
		for idx, cond := range conditions {
			if idx > 0 {
				buffer.WriteString(cs.String())
			}
			unixNano := time.Now().UnixNano()
			if isNull == cond.Symbol || isNotNull == cond.Symbol || isEmpty == cond.Symbol || isNotEmpty == cond.Symbol {
				buffer.WriteString(fmt.Sprintf(cond.Symbol.String(), cond.Colum))
			} else {
				condVar := strconv.FormatInt(unixNano, 10) + "_" + cond.Colum
				condVal[condVar] = cond.Value
				buffer.WriteString(fmt.Sprintf(cond.Symbol.String(), cond.Colum, ":"+condVar))
			}
		}
	}
	ands := bp.ands
	flag := buffer.String() != ""
	handleBracketCondition(ands, And, flag, buffer, condVal)

	ors := bp.ors
	flag = buffer.String() != ""
	handleBracketCondition(ors, Or, flag, buffer, condVal)

	return buffer.String(), condVal
}

func (bp BizParam) getOrderBy() string {
	var buffer bytes.Buffer
	buffer.WriteString(" order by ")

	orderBys := bp.orderBys
	if nil == orderBys || len(orderBys) == 0 {
		buffer.WriteString(" id desc")
		return buffer.String()
	}

	for idx, ob := range orderBys {
		if idx > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(fmt.Sprintf(ob.Sort.String(), ob.Column))
	}
	return buffer.String()
}

func (bp BizParam) getPage() string {
	var buffer bytes.Buffer
	buffer.WriteString("")
	pl := bp.pageLimit
	if nil != pl {
		buffer.WriteString(" limit ")
		buffer.WriteString(strconv.FormatInt(pl.PageSize, 10))
		buffer.WriteString(" offset ")
		buffer.WriteString(strconv.FormatInt(pl.GetStartRow(), 10))
	}
	return buffer.String()
}

func handleBracketCondition(bps []BizParam, cs ConnSymbol, flag bool, buffer bytes.Buffer, condVal map[string]interface{}) {
	if nil != bps && len(bps) != 0 {
		for _, bp := range bps {
			c, cv := bp.getCondition(cs)
			if c != "" {
				if !flag {
					flag = true
				} else {
					buffer.WriteString(And.String())
				}
				buffer.WriteString(c)
				for k, v := range cv {
					condVal[k] = v
				}
			}
		}
	}
}

func (bp BizParam) addCond(fn string, symbol Symbol, val interface{}) BizParam {
	cond := paramCondition{
		Colum:  fn,
		Symbol: symbol,
		Value:  val,
	}
	bp.condition = append(bp.condition, cond)
	return bp
}

type between struct {
	From interface{}
	To   interface{}
}

type orderBy struct {
	Column string
	Sort   Sort
}

type countResult struct {
	Count int64 `db:"count"`
}

func camelToField(camel string) string {
	if strings.TrimSpace(camel) == "" {
		return ""
	}
	var buffer bytes.Buffer
	runes := []rune(camel)
	for i := 0; i < len(camel); i++ {
		c := runes[i]
		if i != 0 {
			if unicode.IsUpper(c) {
				buffer.WriteString("_")
			}
		}
		buffer.WriteRune(c)
	}

	return strings.ToLower(buffer.String())
}

func getInsertID(rows *sqlx.Rows) int64 {
	var id int64
	if rows.Next() {
		err := rows.Scan(&id)
		if err != nil {
			fmt.Println(err.Error())
			return 0
		}
	} else {
		return 0
	}
	return id
}

type ConnSymbol int

const (
	And ConnSymbol = iota
	Or
)

func (c ConnSymbol) String() string {
	switch c {
	case And:
		return " and "
	case Or:
		return " or "
	default:
		return "Unknown"
	}
}

type Sort int

const (
	Desc Sort = iota
	Asc
)

func (s Sort) String() string {
	switch s {
	case Desc:
		return " %s desc "
	case Asc:
		return " %s asc "
	default:
		return "Unknown"
	}
}

type Symbol int

const (
	equal Symbol = iota
	notEqual
	greaterThan
	greaterThanOrEqual
	lessThan
	lessThanOrEqual
	in
	notIn
	like
	likeLeft
	likeRight
	notLike
	notLikeLeft
	notLikeRight
	isNull
	isNotNull
	isEmpty
	isNotEmpty
	betweenAnd
	contain
	containAny
	notContain
)

func (s Symbol) String() string {
	switch s {
	case equal:
		return " %s = %s "
	case notEqual:
		return " %s <> %s "
	case greaterThan:
		return " %s > %s "
	case greaterThanOrEqual:
		return " %s >= %s "
	case lessThan:
		return " %s < %s "
	case lessThanOrEqual:
		return " %s <= %s "
	case in:
		return " %s in (%s) "
	case notIn:
		return " %s not in (%s) "
	case like:
		return " %s like '%%' || %s || '%%' "
	case likeLeft:
		return " %s like %s || '%%' "
	case likeRight:
		return " %s like '%%' || '%s' "
	case notLike:
		return " %s not like '%%' || %s || '%%' "
	case notLikeLeft:
		return " %s not like %s || '%%' "
	case notLikeRight:
		return " %s not like '%%' || '%s' "
	case isNull:
		return " %s is null "
	case isNotNull:
		return " %s is not null "
	case isEmpty:
		return " %s = '' "
	case isNotEmpty:
		return " %s <> '' "
	case betweenAnd:
		return " %s between %s and %s "
	case contain:
		return " %s @> %s "
	case containAny:
		return " %s && %s "
	case notContain:
		return "NOT(%s && %s)"
	default:
		return "Unknown"
	}
}
