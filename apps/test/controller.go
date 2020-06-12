package test

import (
	"github.com/getevo/evo"
	"github.com/getevo/evo/html"
	"github.com/getevo/evo/lib/fontawesome"
	"github.com/getevo/evo/viewfn"
)

type Controller struct{}

func FilterViewController(r *evo.Request) {
	fv := viewfn.FilterView{
		Model: MyModel{},
		Join: []viewfn.Join{
			{MyGroup{}, "group", "id"},
		},
		Columns: []viewfn.Column{
			{Type: viewfn.None, Title: "ID", Name: "id"},
			{Type: viewfn.TEXT, Title: "Name", Name: "name", SimpleFilter: "MyModel.name LIKE '%*%'"},
			{Type: viewfn.TEXT, Title: "Username", Name: "username", SimpleFilter: "username LIKE '%*%'"},
			{Type: viewfn.SELECT, Select: "group", Title: "Group", Model: MyGroup{}, Name: "name", Alias: "group", Options: []html.KeyValue{
				{1, "Group 1"},
				{2, "Group 2"},
				{3, "Group 3"},
			}, SimpleFilter: "MyModel.\"group\" = '*'",
				Processor: func(column viewfn.Column, data map[string]interface{}, r *evo.Request) string {
					return html.Tag("a", data["group"]).Set("href", "#").Render()
				},
			},

			{Type: viewfn.ACTIONS, Title: "Actions", Options: []html.KeyValue{
				{viewfn.SEARCH, html.Icon(fontawesome.Search)},
				{viewfn.RESET, html.Icon(fontawesome.Undo)},
			},
				Processor: func(column viewfn.Column, data map[string]interface{}, r *evo.Request) string {

					return html.Render(

						[]*html.Element{
							html.Tag("a", "View").Set("class", "btn btn-success").Set("href", "#"),
							html.Tag("a", "Delete").Set("class", "btn btn-danger").Set("onclick", "fv.remove(this)"),
						},
					)

				},
			},
		},
	}

	fv.Prepare(r)
	r.View(fv, "template.filterview", "template.default")

}
