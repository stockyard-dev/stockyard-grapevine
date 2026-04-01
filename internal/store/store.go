package store
import("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{*sql.DB}
type Article struct{ID int64 `json:"id"`;Title string `json:"title"`;Body string `json:"body"`;Category string `json:"category"`;ViewCount int `json:"view_count"`;CreatedAt time.Time `json:"created_at"`;UpdatedAt time.Time `json:"updated_at"`}
func Open(d string)(*DB,error){os.MkdirAll(d,0755);dsn:=filepath.Join(d,"grapevine.db")+"?_journal_mode=WAL&_busy_timeout=5000";db,err:=sql.Open("sqlite",dsn);if err!=nil{return nil,fmt.Errorf("open: %w",err)};db.SetMaxOpenConns(1);migrate(db);return &DB{db},nil}
func migrate(db *sql.DB){db.Exec(`CREATE TABLE IF NOT EXISTS articles(id INTEGER PRIMARY KEY AUTOINCREMENT,title TEXT NOT NULL,body TEXT DEFAULT '',category TEXT DEFAULT 'general',view_count INTEGER DEFAULT 0,created_at DATETIME DEFAULT CURRENT_TIMESTAMP,updated_at DATETIME DEFAULT CURRENT_TIMESTAMP)`)}
func(db *DB)Create(a *Article)error{res,err:=db.Exec(`INSERT INTO articles(title,body,category)VALUES(?,?,?)`,a.Title,a.Body,a.Category);if err!=nil{return err};a.ID,_=res.LastInsertId();return nil}
func(db *DB)List(q,category string)([]Article,error){base:=`SELECT id,title,body,category,view_count,created_at,updated_at FROM articles WHERE 1=1`;args:=[]interface{}{};if q!=""{base+=` AND (title LIKE ? OR body LIKE ?)`;args=append(args,"%"+q+"%","%"+q+"%")};if category!=""{base+=` AND category=?`;args=append(args,category)};base+=` ORDER BY view_count DESC,updated_at DESC`;rows,err:=db.Query(base,args...);if err!=nil{return nil,err};defer rows.Close();var out[]Article;for rows.Next(){var a Article;rows.Scan(&a.ID,&a.Title,&a.Body,&a.Category,&a.ViewCount,&a.CreatedAt,&a.UpdatedAt);out=append(out,a)};return out,nil}
func(db *DB)View(id int64){db.Exec(`UPDATE articles SET view_count=view_count+1 WHERE id=?`,id)}
func(db *DB)Update(id int64,title,body,category string){db.Exec(`UPDATE articles SET title=?,body=?,category=?,updated_at=CURRENT_TIMESTAMP WHERE id=?`,title,body,category,id)}
func(db *DB)Delete(id int64){db.Exec(`DELETE FROM articles WHERE id=?`,id)}
func(db *DB)Stats()(map[string]interface{},error){var total int;db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&total);return map[string]interface{}{"articles":total},nil}
