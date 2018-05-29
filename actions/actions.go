package actions

import "github.com/tochk/kirino/server"

func SaveAction(userName, ip, action, id, item string) error {
	_, err := server.Core.Db.Query("INSERT INTO actions (username, ip, action, itemid, item) VALUES (?, ?, ?, ?, ?)", userName, ip, action, id, item)
	return err
}

//todo view actions
