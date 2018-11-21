package abstract_sql

import (
	"database/sql"
	"fmt"

	"github.com/draleyva/seaweedfs/weed/filer2"
	"github.com/draleyva/seaweedfs/weed/glog"
)

type AbstractSqlStore struct {
	DB               *sql.DB
	SqlInsert        string
	SqlUpdate        string
	SqlFind          string
	SqlDelete        string
	SqlListExclusive string
	SqlListInclusive string
}

func (store *AbstractSqlStore) InsertEntry(entry *filer2.Entry) (err error) {

	dir, name := entry.FullPath.DirAndName()
	meta, err := entry.EncodeAttributesAndChunks()
	if err != nil {
		return fmt.Errorf("encode %s: %s", entry.FullPath, err)
	}

	res, err := store.DB.Exec(store.SqlInsert, hashToLong(dir), name, dir, meta)
	if err != nil {
		return fmt.Errorf("insert %s: %s", entry.FullPath, err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("insert %s but no rows affected: %s", entry.FullPath, err)
	}
	return nil
}

func (store *AbstractSqlStore) UpdateEntry(entry *filer2.Entry) (err error) {

	dir, name := entry.FullPath.DirAndName()
	meta, err := entry.EncodeAttributesAndChunks()
	if err != nil {
		return fmt.Errorf("encode %s: %s", entry.FullPath, err)
	}

	res, err := store.DB.Exec(store.SqlUpdate, meta, hashToLong(dir), name, dir)
	if err != nil {
		return fmt.Errorf("update %s: %s", entry.FullPath, err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update %s but no rows affected: %s", entry.FullPath, err)
	}
	return nil
}

func (store *AbstractSqlStore) FindEntry(fullpath filer2.FullPath) (*filer2.Entry, error) {

	dir, name := fullpath.DirAndName()
	row := store.DB.QueryRow(store.SqlFind, hashToLong(dir), name, dir)
	var data []byte
	if err := row.Scan(&data); err != nil {
		return nil, filer2.ErrNotFound
	}

	entry := &filer2.Entry{
		FullPath: fullpath,
	}
	if err := entry.DecodeAttributesAndChunks(data); err != nil {
		return entry, fmt.Errorf("decode %s : %v", entry.FullPath, err)
	}

	return entry, nil
}

func (store *AbstractSqlStore) DeleteEntry(fullpath filer2.FullPath) error {

	dir, name := fullpath.DirAndName()

	res, err := store.DB.Exec(store.SqlDelete, hashToLong(dir), name, dir)
	if err != nil {
		return fmt.Errorf("delete %s: %s", fullpath, err)
	}

	_, err = res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete %s but no rows affected: %s", fullpath, err)
	}

	return nil
}

func (store *AbstractSqlStore) ListDirectoryEntries(fullpath filer2.FullPath, startFileName string, inclusive bool, limit int) (entries []*filer2.Entry, err error) {

	sqlText := store.SqlListExclusive
	if inclusive {
		sqlText = store.SqlListInclusive
	}

	rows, err := store.DB.Query(sqlText, hashToLong(string(fullpath)), startFileName, string(fullpath), limit)
	if err != nil {
		return nil, fmt.Errorf("list %s : %v", fullpath, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var data []byte
		if err = rows.Scan(&name, &data); err != nil {
			glog.V(0).Infof("scan %s : %v", fullpath, err)
			return nil, fmt.Errorf("scan %s: %v", fullpath, err)
		}

		entry := &filer2.Entry{
			FullPath: filer2.NewFullPath(string(fullpath), name),
		}
		if err = entry.DecodeAttributesAndChunks(data); err != nil {
			glog.V(0).Infof("scan decode %s : %v", entry.FullPath, err)
			return nil, fmt.Errorf("scan decode %s : %v", entry.FullPath, err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
