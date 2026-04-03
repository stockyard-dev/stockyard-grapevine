package main
import ("fmt";"log";"net/http";"os";"github.com/stockyard-dev/stockyard-grapevine/internal/server";"github.com/stockyard-dev/stockyard-grapevine/internal/store")
func main(){port:=os.Getenv("PORT");if port==""{port="9700"};dataDir:=os.Getenv("DATA_DIR");if dataDir==""{dataDir="./grapevine-data"}
db,err:=store.Open(dataDir);if err!=nil{log.Fatalf("grapevine: %v",err)};defer db.Close();srv:=server.New(db,server.DefaultLimits())
fmt.Printf("\n  Grapevine — Self-hosted knowledge base and FAQ\n  Dashboard:  http://localhost:%s/ui\n  API:        http://localhost:%s/api\n\n",port,port)
log.Printf("grapevine: listening on :%s",port);log.Fatal(http.ListenAndServe(":"+port,srv))}
