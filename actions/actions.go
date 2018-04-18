package actions

func (s *server) saveAction(userName, ip, action, id, item string) error {
	_, err := s.Db.Query("INSERT INTO actions (username, ip, action, itemid, item) VALUES (?, ?, ?, ?, ?)", userName, ip, action, id, item)
	return err
}