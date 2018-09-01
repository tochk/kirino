package pagination

import (
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func Calc(page, count int) (pagination html.Pagination) {
	if page < 1 {
		page = 1
	}
	pagination.CurrentPage = page
	pagination.PerPage = server.Config.PerPage
	pagination.Offset = server.Config.PerPage * (page - 1)

	if count > server.Config.PerPage*page {
		pagination.NextPage = pagination.CurrentPage + 1
		if pagination.NextPage != (count/server.Config.PerPage)+1 {
			pagination.LastPage = (count / server.Config.PerPage) + 1
		}
	}
	if pagination.CurrentPage > 1 {
		pagination.PrevPage = pagination.CurrentPage - 1
	}
	return pagination
}
