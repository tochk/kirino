package pagination

import (
	"github.com/tochk/kirino_wifi/templates/html"
)

type Pagination = html.Pagination

func Calc(page, count int) (pagination Pagination) {
	if page < 1 {
		page = 1
	}
	pagination.CurrentPage = page
	pagination.PerPage = perPage
	pagination.Offset = perPage * (page - 1)

	if count > perPage*page {
		pagination.NextPage = pagination.CurrentPage + 1
		if pagination.NextPage != (count/perPage)+1 {
			pagination.LastPage = (count / perPage) + 1
		}
	}
	if pagination.CurrentPage > 1 {
		pagination.PrevPage = pagination.CurrentPage - 1
	}
	return pagination
}
