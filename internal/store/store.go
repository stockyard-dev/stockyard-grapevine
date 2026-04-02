package store
import ("database/sql";"fmt";"os";"path/filepath";"time";_ "modernc.org/sqlite")
type DB struct{db *sql.DB}
type Article struct {
	ID string `json:"id"`
	Title string `json:"title"`
	Body string `json:"body"`
	Category string `json:"category"`
	Tags string `json:"tags"`
	Slug string `json:"slug"`
	Helpful int `json:"helpful"`
	NotHelpful int `json:"not_helpful"`
	Status string `json:"status"`
	CreatedAt string `json:"created_at"`
}
func Open(d string)(*DB,error){if err:=os.MkdirAll(d,0755);err!=nil{return nil,err};db,err:=sql.Open("sqlite",filepath.Join(d,"grapevine.db")+"?_journal_mode=WAL&_busy_timeout=5000");if err!=nil{return nil,err}
db.Exec(`CREATE TABLE IF NOT EXISTS articles(id TEXT PRIMARY KEY,title TEXT NOT NULL,body TEXT DEFAULT '',category TEXT DEFAULT '',tags TEXT DEFAULT '',slug TEXT DEFAULT '',helpful INTEGER DEFAULT 0,not_helpful INTEGER DEFAULT 0,status TEXT DEFAULT 'published',created_at TEXT DEFAULT(datetime('now')))`)
return &DB{db:db},nil}
func(d *DB)Close()error{return d.db.Close()}
func genID()string{return fmt.Sprintf("%d",time.Now().UnixNano())}
func now()string{return time.Now().UTC().Format(time.RFC3339)}
func(d *DB)Create(e *Article)error{e.ID=genID();e.CreatedAt=now();_,err:=d.db.Exec(`INSERT INTO articles(id,title,body,category,tags,slug,helpful,not_helpful,status,created_at)VALUES(?,?,?,?,?,?,?,?,?,?)`,e.ID,e.Title,e.Body,e.Category,e.Tags,e.Slug,e.Helpful,e.NotHelpful,e.Status,e.CreatedAt);return err}
func(d *DB)Get(id string)*Article{var e Article;if d.db.QueryRow(`SELECT id,title,body,category,tags,slug,helpful,not_helpful,status,created_at FROM articles WHERE id=?`,id).Scan(&e.ID,&e.Title,&e.Body,&e.Category,&e.Tags,&e.Slug,&e.Helpful,&e.NotHelpful,&e.Status,&e.CreatedAt)!=nil{return nil};return &e}
func(d *DB)List()[]Article{rows,_:=d.db.Query(`SELECT id,title,body,category,tags,slug,helpful,not_helpful,status,created_at FROM articles ORDER BY created_at DESC`);if rows==nil{return nil};defer rows.Close();var o []Article;for rows.Next(){var e Article;rows.Scan(&e.ID,&e.Title,&e.Body,&e.Category,&e.Tags,&e.Slug,&e.Helpful,&e.NotHelpful,&e.Status,&e.CreatedAt);o=append(o,e)};return o}
func(d *DB)Delete(id string)error{_,err:=d.db.Exec(`DELETE FROM articles WHERE id=?`,id);return err}
func(d *DB)Count()int{var n int;d.db.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&n);return n}
