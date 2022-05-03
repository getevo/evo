package viewfn

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/getevo/evo"
	"github.com/getevo/evo/html"
	"github.com/getevo/evo/lib/T"
	"github.com/getevo/evo/menu"
)

type ColumnType int
type ActionType string

const (
	RESET    ActionType = "reset"
	SEARCH   ActionType = "search"
	PAGESIZE ActionType = "pagesize"
	ORDER    ActionType = "order"
)

const (
	TEXT    ColumnType = 0
	NUMBER  ColumnType = 1
	DATE    ColumnType = 2
	HTML    ColumnType = 3
	RANGE   ColumnType = 4
	SELECT  ColumnType = 5
	CUSTOM  ColumnType = 6
	ACTIONS ColumnType = 7
	None    ColumnType = 8
)

type Join struct {
	Model  interface{}
	MainFK string
	DestFK string
}

type Column struct {
	Type         ColumnType
	Title        string
	Width        string
	Resize       bool
	Order        bool
	Name         string
	Alias        string
	Select       string
	Options      []html.KeyValue
	Actions      html.Renderable
	InputBuilder func(r *evo.Request) html.Renderable
	Attribs      html.Attributes
	QueryBuilder func(r *evo.Request) []string
	SimpleFilter string
	Processor    func(column Column, data map[string]interface{}, r *evo.Request) string
	Model        interface{}
}

type Pagination struct {
	Records     int
	CurrentPage int
	Pages       int
	Limit       int
	First       int
	Last        int
	PageRange   []int
}

type Sort struct {
	PrimaryKey string
	SortColumn string
}
type FilterView struct {
	Sort        Sort
	Title       string
	Description string
	Style       string
	Columns     []Column
	Select      []string
	Model       interface{}
	Entity      string
	Join        []Join
	Attribs     html.Attributes
	Unscoped    bool
	PickerMode  bool
	PickerTitle func(data map[string]interface{}, r *evo.Request) string
	PickerID    func(data map[string]interface{}, r *evo.Request) string

	QueryBuilder func(r *evo.Request) []string
	data         []map[string]interface{}
	Pagination   Pagination
	EnableDebug  bool
	PageActions  []menu.Menu
	BatchActions []menu.Menu
}

func (fv *FilterView) Debug() *FilterView {
	fv.EnableDebug = true
	return fv
}
func (fv FilterView) GetData() []map[string]interface{} {
	return fv.data
}

func getName(t reflect.Type) string {
	parts := strings.Split(t.Name(), ".")
	return parts[len(parts)-1]
}

func quote(s string) string {
	return "`" + s + "`"
}

func defaultProcessor(column Column, data map[string]interface{}, r *evo.Request) string {
	if column.Alias == "" {
		if v, ok := data[column.Name]; ok {
			return fmt.Sprint(v)
		}
	} else {
		if v, ok := data[column.Alias]; ok {
			return fmt.Sprint(v)
		}
	}
	return ""
}

func (fv *FilterView) Prepare(r *evo.Request) bool {
	var db = evo.GetDBO()
	var query = []string{"true"}
	var _select = fv.Select
	var _join string
	var models = map[string]string{}
	var tables []string
	fv.Pagination.Limit = 10
	var offset = 0
	var order = ""
	if r.Query("size") != "" {
		fv.Pagination.Limit = T.Must(r.Query("size")).Int()
		if fv.Pagination.Limit < 10 {
			fv.Pagination.Limit = 10
		}
		if fv.Pagination.Limit > 100 {
			fv.Pagination.Limit = 100
		}
	} else if r.Cookies("tableSize") != "" {
		fv.Pagination.Limit = T.Must(r.Cookies("tableSize")).Int()
		if fv.Pagination.Limit < 10 {
			fv.Pagination.Limit = 10
		}
		if fv.Pagination.Limit > 100 {
			fv.Pagination.Limit = 100
		}
	}
	if r.Query("page") != "" {
		fv.Pagination.CurrentPage = T.Must(r.Query("page")).Int() - 1
		offset = fv.Pagination.CurrentPage * fv.Pagination.Limit
		if offset < 0 {
			offset = 0
		}
	}

	db = evo.GetDBO()
	var schema = db.Model(fv.Model).Statement
	schema.Parse(fv.Model)

	for _, field := range schema.Schema.Fields {
		if tag := field.Tag.Get("fv"); tag != "" {
			if strings.Contains(tag, "orderable") {
				fv.Sort = Sort{
					SortColumn: field.DBName,
					PrimaryKey: schema.Schema.PrioritizedPrimaryField.DBName,
				}
			}
			break
		}
	}

	s1 := r.Query("order")
	if s1 != "" {
		ok := false
		if s1 == fv.Sort.SortColumn {
			ok = true
		} else {
			for _, column := range fv.Columns {
				if s1 == column.Name {
					ok = true
				}
			}
		}
		if ok {
			s2 := strings.ToUpper(r.Query("sort"))
			if s2 == "ASC" || s2 == "DESC" {
				order = "`" + s1 + "` " + s2
			} else {
				order = "`" + s1 + "` ASC"
			}
		}

	}

	tables = append(tables, schema.Table)
	models[getName(reflect.TypeOf(fv.Model))] = tables[0]
	for _, join := range fv.Join {
		var schema = db.Model(join.Model).Statement
		schema.Parse(join.Model)
		t := schema.Table
		models[getName(reflect.TypeOf(join.Model))] = t
		tables = append(tables, t)
		_join += " INNER JOIN " + quote(t) + " ON " + quote(tables[0]) + "." + quote(join.MainFK) + " = " + quote(t) + "." + quote(join.DestFK)
	}

	if fv.QueryBuilder != nil {
		query = append(query, fv.QueryBuilder(r)...)
	}
	for k, column := range fv.Columns {
		if column.Model == nil {
			column.Model = quote(tables[0])
		} else {
			if _, ok := column.Model.(string); !ok {
				var schema = db.Model(column.Model).Statement
				schema.Parse(column.Model)
				column.Model = quote(schema.Table)
			}
		}
		if column.Alias == "" {
			column.Alias = column.Name
		}

		if column.Processor == nil {
			fv.Columns[k].Processor = defaultProcessor
		}

		if column.Select != "" && column.Select != "-" {
			_select = append(_select, column.Select+" AS "+quote(column.Alias))
		} else if column.Select != "-" {
			if column.Name != "" && column.Name != "-" {
				_select = append(_select, column.Model.(string)+"."+quote(column.Name)+" AS "+quote(column.Alias))
			}
		}

		if column.QueryBuilder != nil {
			query = append(query, column.QueryBuilder(r)...)
		} else {

			var v string
			if column.Name != "" {
				v = r.Query(column.Name)
			} else {
				v = r.Query(column.Alias)
			}
			if v != "" {
				q := column.SimpleFilter
				for model, tb := range models {
					q = strings.Replace(q, model+".", quote(tb)+".", -1)
				}
				q = strings.Replace(q, "*", v, -1)

				query = append(query, q)
			}
		}

	}

	if order == "" {
		if fv.Sort.SortColumn != "" {
			order = quote(tables[0]) + "." + quote(fv.Sort.SortColumn) + " DESC"
		} else if len(schema.Schema.PrimaryFieldDBNames) > 0 {
			order = quote(tables[0]) + "." + quote(schema.Schema.PrimaryFieldDBNames[0]) + " DESC"
		}
	}

	if fv.Unscoped {
		db = db.Unscoped()
	}
	if fv.Sort.SortColumn != "" {
		_select = append(_select, quote(tables[0])+"."+quote(fv.Sort.SortColumn)+" AS `_order`")
	}
	_select = append(_select, quote(tables[0])+"."+quote(schema.Schema.PrimaryFieldDBNames[0])+" AS `pk`")
	var dataQuery = ""
	if fv.PickerMode && r.Query("pk") != "" {
		dataQuery = fmt.Sprintf("SELECT %s FROM %s %s WHERE "+quote(tables[0])+"."+quote(schema.Schema.PrimaryFieldDBNames[0])+" = %s",
			strings.Join(_select, ","),
			quote(tables[0]), //main table
			_join,
			strconv.Quote(r.Query("pk")),
		)
	} else {
		dataQuery = fmt.Sprintf("SELECT %s FROM %s %s WHERE %s ORDER BY %s LIMIT %d OFFSET %d ",
			strings.Join(_select, ","),
			quote(tables[0]), //main table
			_join,
			strings.Join(query, " AND "),
			order,
			fv.Pagination.Limit,
			offset,
		)
	}

	limitQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s WHERE %s",
		quote(tables[0]), //main table
		_join,
		strings.Join(query, " AND "),
	)
	db = evo.GetDBO()
	if fv.EnableDebug {
		db = db.Debug()
	}
	if !fv.PickerMode && r.Query("pk") == "" {
		row := db.Raw(limitQuery).Row()
		row.Scan(&fv.Pagination.Records)
		fv.Pagination.Pages = (fv.Pagination.Records / fv.Pagination.Limit) + 1
		fv.Pagination.First = fv.Pagination.CurrentPage * fv.Pagination.Limit
		fv.Pagination.Last = fv.Pagination.First + fv.Pagination.Limit
		if fv.Pagination.Last > fv.Pagination.Records {
			fv.Pagination.Last = fv.Pagination.Records
		}
		to := fv.Pagination.CurrentPage + 5
		if to > fv.Pagination.Pages {
			to = fv.Pagination.Pages + 1
		}
		for i := fv.Pagination.CurrentPage - 2; i < to; i++ {
			if i > 0 {
				fv.Pagination.PageRange = append(fv.Pagination.PageRange, i)
			}
		}
	}

	rows, err := db.Raw(dataQuery).Rows()

	if err != nil {
		return false
	}
	columns, err := rows.Columns()
	if err != nil {
		return false
	}
	length := len(columns)
	fv.data = make([]map[string]interface{}, 0)
	for rows.Next() {

		current := makeResultReceiver(length)
		if err := rows.Scan(current...); err != nil {
			panic(err)
		}
		value := make(map[string]interface{})
		for i := 0; i < length; i++ {
			k := columns[i]
			v := reflect.ValueOf(current[i]).Elem().Interface()
			value[k] = byteToStr(v)
		}
		if fv.PickerID != nil {
			value["_picker_pk"] = fv.PickerID(value, r)
		} else {
			value["_picker_pk"] = value["pk"]
		}

		if fv.PickerTitle != nil {
			value["_picker_title"] = fv.PickerTitle(value, r)
		} else {
			value["_picker_title"] = ""
		}
		fv.data = append(fv.data, value)
	}
	if fv.PickerMode && r.Query("pk") != "" {
		var title = ""
		if fv.PickerTitle != nil {
			title = fv.PickerTitle(fv.data[0], r)
		}
		var pk = fv.data[0]["pk"]
		if fv.PickerID != nil {
			pk = fv.PickerID(fv.data[0], r)
		}
		r.WriteResponse(map[string]interface{}{
			"pk":    pk,
			"title": title,
		})
		return true
	}
	return false
}

func makeResultReceiver(length int) []interface{} {
	result := make([]interface{}, 0, length)
	for i := 0; i < length; i++ {
		var current interface{}
		current = struct{}{}
		result = append(result, &current)
	}
	return result
}

func byteToStr(v interface{}) string {

	if cast, ok := v.([]byte); ok {
		return string(cast)
	} else if v == nil {
		return ""
	} else {
		return fmt.Sprintf("%v", v)
	}

}

func (fv FilterView) SizeInput(r *evo.Request) string {

	return html.Render(
		html.Tag("div", html.Input("select", "size", "").SetOptions([]html.KeyValue{
			{10, "10"},
			{25, "25"},
			{50, "50"},
			{100, "100"},
		}).SetAttr("onchange", "fv.setSize(this)").SetValue(r.Query("size"))).Set("class", "fv-action-pagesize"))
}

func (col Column) Filter(r *evo.Request) html.Renderable {
	if col.InputBuilder != nil {
		return col.InputBuilder(r)
	}

	var el *html.InputStruct
	switch col.Type {
	case None:
		return html.Tag("div", "")
		break
	case NUMBER:
		el = html.Input("number", col.Name, "")
		el.SetAttr("onpressenter", "fv.filter(this)")
		break
	case DATE:
		el = html.Input("daterange", col.Name, "")
		el.SetAttr("onpressenter", "fv.filter(this)")
		break
	case SELECT:
		el = html.Input("select", col.Name, "").SetOptions(col.Options)
		el.SetAttr("onchange", "fv.filter(this)")
		el.Options = append([]html.KeyValue{
			{"", "--------"},
		}, el.Options...)
	case ACTIONS:
		if col.Actions != nil {
			return col.Actions
		}
		var actions []html.Element
		for _, item := range col.Options {
			action := item.Key.(ActionType)
			switch action {
			case SEARCH:
				actions = append(actions, *html.Tag("button", item.Value).Set("class", "btn fv-action-btn").Set("onclick", "fv.apply(this)"))
				break
			case RESET:
				actions = append(actions, *html.Tag("button", item.Value).Set("class", "btn fv-action-btn").Set("onclick", "fv.reset(this)"))
				break
			case PAGESIZE:
				actions = append(actions, *html.Tag("div", html.Input("select", "size", fmt.Sprint(item.Value)).SetAttr("onchange", "fv.setSize(this)").SetOptions([]html.KeyValue{
					{10, "10"},
					{25, "25"},
					{50, "50"},
					{100, "100"},
				})).Set("class", "fv-action-pagesize"))
				break
			}

		}
		return html.Tag("div", actions).Set("class", "fv-actions")
	default:
		el = html.Input("text", col.Name, "")
		el.SetAttr("onpressenter", "fv.filter(this)")
	}
	if col.Name == "" && col.Select != "" {
		el.SetName(col.Select)
	}

	if col.Attribs != nil {
		el.Attributes = col.Attribs
	}
	if col.Title != "" {
		el.Placeholder(col.Title)
	}
	if r.Query(col.Name) != "" {
		el.Value = r.Query(col.Name)
	} else if r.Query(col.Name) != "" {
		el.Value = r.Query(col.Name)
	} else {
		el.Value = ""
	}
	return el
}
