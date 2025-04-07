package main

import (
	"context"
	"log"
	"mime"
	"net/http"

	"git.divar.cloud/divar/girls-hackathon/realestate-poi/internal/handlers"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/internal/services"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/configs"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/database"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/database/db"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/pkg/transport"
	"git.divar.cloud/divar/girls-hackathon/realestate-poi/utils"
	_ "github.com/golang-migrate/migrate/v4/database/postgres" // Postgres driver
	_ "github.com/golang-migrate/migrate/v4/source/file"       //  'file' source driver
	"github.com/gorilla/mux"
)

func main() {

	conf, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load configurations: %s\n", err)
	}

	conPool, err := database.ConnectToDatabase(
		context.Background(),
		conf.Database.Username,
		conf.Database.Password,
		conf.Database.Host,
		conf.Database.DBName,
		conf.Database.Port,
		conf.Database.SSLMode,
		conf.Database.MaxConns,
		conf.Database.MinConns,
		conf.Database.MaxConnLifetimeJitterMinutes,
		conf.Database.MaxConnLifetimeMinutes,
		conf.Database.MaxConnIdleTimeMinutes,
	)
	if err != nil {
		log.Fatal("Error in databse setup " + err.Error())
	}

	query := db.New(conPool)
	sessionStore := utils.NewSessionStore(&conf.Session)
	// fmt.Println(conf.Session.AuthKey)
	// fmt.Println(reflect.ValueOf(conf.Session.AuthKey).Kind())
	jwtManager := utils.NewJWTManager(&conf.Jwt)

	snapp := transport.NewSnapp(&conf.Snapp)
	// log.Println(conf.Snapp.ApiKey)
	tapsi := transport.NewTapsi(&conf.Tapsi)
	neshan := transport.NewNeshan(&conf.Neshan)
	taxiService := services.NewTransportService(snapp, tapsi, neshan, query)

	kenarService := services.NewKenarService(conf.Kenar.ApiKey, "https://api.divar.ir/v1/open-platform", query)
	kenarHandler := handlers.NewKenarHandler(sessionStore, kenarService, taxiService)
	pageHandler := handlers.NewPageHandler(sessionStore, kenarService, taxiService)
	oauthService := services.NewOAuthService(conf.Kenar, query, conPool)
	oauthHandler := handlers.NewOAuthHandler(sessionStore, oauthService, kenarService, jwtManager)
	// oauthHandler := handlers.NewOAuthHandler(oauthService)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	r.HandleFunc("/poi", kenarHandler.Poi)
	r.HandleFunc("/addon/oauth", oauthHandler.AddonOauth)
	r.HandleFunc("/api/calculate-fare", kenarHandler.GetPrice)
	r.HandleFunc("/api/find-amenities", jwtManager.JWTMiddlewear(kenarHandler.Poi))
	r.HandleFunc("/api/record-purchase", jwtManager.JWTMiddlewear(kenarHandler.RecordPurchase))
	r.HandleFunc("/api/add-to-ad", jwtManager.JWTMiddlewear(kenarHandler.AddLocationWidget))

	r.HandleFunc("/api/get-origin", kenarHandler.GetOriginCoordinates).Methods("POST")
	r.HandleFunc("/oauth/callback", oauthHandler.OauthCallback)

	//front end
	// Serve static files (CSS, JS)
	mime.AddExtensionType(".css", "text/css")

	// fileServer := http.FileServer(http.Dir("./web/static"))
	// r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fileServer))
	r.HandleFunc("/api/seller/landing", pageHandler.SellerDashboardHandler).Methods("GET")
	r.Handle("/api/buyer/landing", jwtManager.JWTMiddlewear(pageHandler.BuyerDashboardHandler)).Methods("GET")
	htmlFileServer := http.FileServer(http.Dir("./web"))
	r.PathPrefix("/web/").Handler(http.StripPrefix("/web/", htmlFileServer))
	r.HandleFunc("/error", utils.RenderErrorPage)

	port := conf.Server.Port
	log.Printf("Server started on port %s", port)
	http.ListenAndServe(":"+port, r)
}
