package pagination

import "log"

//todo rewrite
func (s *server) paginationCalc(page, perPage int, table string) Pagination {
	var (
		count      int
		pagination Pagination
		err        error
	)
	if page < 1 {
		page = 1
	}
	pagination.CurrentPage = page
	pagination.PerPage = perPage
	pagination.Offset = perPage * (page - 1)
	switch table {
	case "wifiUsers":
		err = s.Db.Get(&count, "SELECT COUNT(*) as count FROM wifiUsers")
	case "memorandums":
		err = s.Db.Get(&count, "SELECT COUNT(*) as count FROM memorandums")
	case "departments":
		err = s.Db.Get(&count, "SELECT COUNT(*) as count FROM departments")
	}
	if err != nil {
		log.Println(err)
		return Pagination{}
	}
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
