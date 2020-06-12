package viewfn

import (
	"fmt"
	"github.com/getevo/evo"
	"github.com/getevo/evo/html"
	"github.com/getevo/evo/lib/T"
	"reflect"
	"strings"
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
	Width        int
	Resize       bool
	Order        bool
	Name         string
	Alias        string
	Select       string
	Options      []html.KeyValue
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

type FilterView struct {
	Style        string
	Columns      []Column
	Model        interface{}
	Join         []Join
	Attribs      html.Attributes
	Unscoped     bool
	QueryBuilder func(r *evo.Request) []string
	data         []map[string]interface{}
	Pagination   Pagination
}

func (fv FilterView) GetData() []map[string]interface{} {
	return fv.data
}

func getName(t reflect.Type) string {
	parts := strings.Split(t.Name(), ".")
	return parts[len(parts)-1]
}

func quote(s string) string {
	return "\"" + s + "\""
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

func (fv *FilterView) Prepare(r *evo.Request) {
	var db = evo.GetDBO()
	var query = []string{"true"}
	var _select []string
	var _join string
	var models = map[string]string{}
	var tables []string
	fv.Pagination.Limit = 10
	var offset = 0
	var order = ""
	if r.Query("limit") != "" {
		fv.Pagination.Limit = T.Must(r.Query("limit")).Int()
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
	s1 := r.Query("order")
	if s1 != "" {

		ok := false
		for _, column := range fv.Columns {
			if s1 == column.Name {
				ok = true
			}
		}
		if ok {
			s2 := strings.ToUpper(r.Query("sort"))
			if s2 == "ASC" || s2 == "DESC" {
				order = s1 + " " + s2
			} else {
				order = s1 + " ASC"
			}
		}

	}
	tables = append(tables, db.NewScope(fv.Model).TableName())
	models[getName(reflect.TypeOf(fv.Model))] = tables[0]
	for _, join := range fv.Join {
		t := db.NewScope(join.Model).TableName()
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
				column.Model = quote(db.NewScope(column.Model).TableName())
			}
		}
		if column.Alias == "" {
			column.Alias = column.Name
		}

		if column.Processor == nil {
			fv.Columns[k].Processor = defaultProcessor
		}

		if column.Name != "" {
			_select = append(_select, column.Model.(string)+"."+quote(column.Name)+" AS "+quote(column.Alias))
		}

		if column.QueryBuilder != nil {
			query = append(query, column.QueryBuilder(r)...)
		} else {
			var v string
			if column.Select != "" {
				v = r.Query(column.Select)
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
		order = quote(tables[0]) + ".\"id\" DESC"
	}
	if fv.Unscoped {
		db = db.Unscoped()
	}

	dataQuery := fmt.Sprintf("SELECT %s FROM %s %s WHERE %s ORDER BY %s LIMIT %d OFFSET %d ",
		strings.Join(_select, ","),
		quote(tables[0]), //main table
		_join,
		strings.Join(query, " AND "),
		order,
		fv.Pagination.Limit,
		offset,
	)

	fmt.Println(dataQuery)
	limitQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s %s WHERE %s",
		quote(tables[0]), //main table
		_join,
		strings.Join(query, " AND "),
	)
	row := db.Raw(limitQuery).Row()
	row.Scan(&fv.Pagination.Records)
	fv.Pagination.Pages = fv.Pagination.Records / fv.Pagination.Limit
	if fv.Pagination.Pages == 0 {
		fv.Pagination.Pages = 1
	}
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

	rows, err := db.Raw(dataQuery).Rows()
	if err != nil {
		return
	}
	columns, err := rows.Columns()
	if err != nil {
		return
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
			value[k] = v
		}
		fv.data = append(fv.data, value)
	}

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
	if col.Select != "" {
		el.SetName(col.Select)
	}

	if col.Attribs != nil {
		el.Attributes = col.Attribs
	}
	if col.Title != "" {
		el.Placeholder(col.Title)
	}
	if r.Query(col.Select) != "" {
		el.Value = r.Query(col.Select)
	} else if r.Query(col.Name) != "" {
		el.Value = r.Query(col.Name)
	} else {
		el.Value = ""
	}
	return el
}
