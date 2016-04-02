package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"

	"labix.org/v2/mgo/bson"

	"github.com/drone/routes"
	"github.com/mkilling/goejdb"
	"github.com/pelletier/go-toml"
)

type Box struct {
	Email           string `json:"email"`
	Zip             string `json:"zip"`
	Country         string `json:"country"`
	Profession      string `json:"profession"`
	Favourite_color string `json:"favourite_color"`
	Is_smoking      string `json:"is_smoking"`
	Favorite_sport  string `json:"favorite_sport"`
	Food            `json:"food"`
	Music           `json:"music"`
	Movie           `json:"movie"`
	Travel          `json:"travel"`
}

type Food struct {
	Type          string `json:"type"`
	Drink_alcohol string `json:"drink_alcohol"`
}

type Music struct {
	Spotify_user_id string `json:"spotify_user_id"`
}

type Movie struct {
	Tv_shows []string `json:"tv_shows"`
	Movies   []string `json:"movies"`
}

type Travel struct {
	Flight `json:"flight"`
}

type Flight struct {
	Seat string `json:"seat"`
}

var dbfile string
var port int64
var rpcServerPort int64
var replica []string

func main() {
	flag.Parse()
	configFile := flag.Arg(0)
	config, err := toml.LoadFile(configFile)
	if err != nil {
		fmt.Println("Error ", err.Error())
		return
	} else {
		// retrieve data directly
		dbfile = config.Get("database.file_name").(string)
		port = config.Get("database.port_num").(int64)

		// or using an intermediate object
		rpcServerPort = config.Get("replication.rpc_server_port_num").(int64)
		replicaConfig := config.Get("replication.replica").([]interface{})
		for _, r := range replicaConfig {
			replica = append(replica, fmt.Sprintf("%v", r))
		}

		fmt.Println("dbfile: ", dbfile)
		fmt.Println("port: ", port)
		fmt.Println("rpc server port: ", rpcServerPort)
		// show where elements are in the file
		fmt.Println("replica: ", replica)
	}
	CreateDb()
	mux := routes.New()
	mux.Get("/profile/:email", GetProfile)
	mux.Post("/profile", PostProfile)
	mux.Del("/profile/:email", DeleteProfile)
	mux.Put("/profile/:email", PutProfile)

	http.Handle("/", mux)
	log.Println("Listening...")
	go ReceiveRPCMsg()
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)

}

var profileitems map[string][]string //travel
var TVprofileitems map[string][]string
var Moviesprofileitems map[string][]string
var counter = 1

func GetProfile(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	email := params.Get(":email")

	var box Box
	var ok bool
	if box, ok = DbGetProfile(email); !ok {
		http.Error(w, fmt.Sprintf("Requested email (%v) is not present", email), 404)
		return
	}

	js, err := json.Marshal(box)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func PostProfile(w http.ResponseWriter, r *http.Request) {

	var u Box
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	if counter == 1 {

		/* create a map*/
		profileitems = make(map[string][]string)
		/* create a map*/
		TVprofileitems = make(map[string][]string)
		/* create a map*/
		Moviesprofileitems = make(map[string][]string)
		counter++

	}

	TVShowitems := u.Movie.Tv_shows
	Moviesitems := u.Movie.Movies

	TVprofileitems[u.Email] = TVShowitems
	Moviesprofileitems[u.Email] = Moviesitems

	/* insert key-value pairs in the map*/
	profileitems[u.Email] = []string{u.Zip, u.Country, u.Profession, u.Favourite_color, u.Is_smoking, u.Favorite_sport, u.Food.Type, u.Food.Drink_alcohol, u.Music.Spotify_user_id, u.Travel.Flight.Seat}

	DbSaveProfile(u)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

func DeleteProfile(w http.ResponseWriter, r *http.Request) {

	params := r.URL.Query()
	email := params.Get(":email")

	delete(profileitems, email)
	delete(TVprofileitems, email)
	delete(Moviesprofileitems, email)

	w.WriteHeader(204)
}
func PutProfile(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	email := params.Get(":email")
	var box Box
	var changed Box
	var i int
	if r.Body == nil {
		http.Error(w, "Please send a request body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(&box)
	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	/* test if entry is present in the database */
	prev, ok := DbGetProfile(email)
	/* if ok is true, entry is present otherwise entry is absent*/
	if !ok {
		log.Println("PUTPROFILE - COULDNT GET PROFILE FOR:", email)
	}
	if ok {
		log.Printf("PUT - new profile: %v\n", box)

		if len(box.Zip) != 0 {
			changed.Zip = box.Zip
		}
		if len(box.Country) != 0 {
			changed.Country = box.Country
		}
		if len(box.Profession) != 0 {
			changed.Profession = box.Profession
		}
		if len(box.Favourite_color) != 0 {
			changed.Favourite_color = box.Favourite_color
		}
		if len(box.Is_smoking) != 0 {
			changed.Is_smoking = box.Is_smoking
		}
		if len(box.Favorite_sport) != 0 {
			changed.Favorite_sport = box.Favorite_sport
		}
		if len(box.Food.Type) != 0 {
			changed.Food.Type = box.Food.Type
		}
		if len(box.Food.Drink_alcohol) != 0 {
			changed.Food.Drink_alcohol = box.Food.Drink_alcohol
		}
		if len(box.Music.Spotify_user_id) != 0 {
			changed.Music.Spotify_user_id = box.Music.Spotify_user_id
		}
		if len(box.Movie.Tv_shows) != 0 {
			changed.Movie.Tv_shows = box.Movie.Tv_shows
		}
		if len(box.Movie.Movies) != 0 {
			changed.Movie.Movies = box.Movie.Movies
		}
		if len(box.Travel.Flight.Seat) != 0 {
			changed.Travel.Flight.Seat = box.Travel.Flight.Seat
		}

		changed.Email = email
		if len(changed.Zip) == 0 {
			changed.Zip = prev.Zip
		}
		if len(changed.Country) == 0 {
			changed.Country = prev.Country
		}
		if len(changed.Profession) == 0 {
			changed.Profession = prev.Profession
		}
		if len(changed.Favourite_color) == 0 {
			changed.Favourite_color = prev.Favourite_color
		}
		if len(changed.Is_smoking) == 0 {
			changed.Is_smoking = prev.Is_smoking
		}
		if len(changed.Favorite_sport) == 0 {
			changed.Favorite_sport = prev.Favorite_sport
		}
		if len(changed.Food.Type) == 0 {
			changed.Food.Type = prev.Food.Type
		}
		if len(changed.Food.Drink_alcohol) == 0 {
			changed.Food.Drink_alcohol = prev.Food.Drink_alcohol
		}
		if len(changed.Music.Spotify_user_id) == 0 {
			changed.Music.Spotify_user_id = prev.Music.Spotify_user_id
		}
		if len(changed.Movie.Tv_shows) == 0 {
			changed.Movie.Tv_shows = []string{}
			for i = 0; i < len(TVprofileitems[email]); i++ {
				changed.Movie.Tv_shows = append(changed.Movie.Tv_shows, TVprofileitems[email][i])
			}
		}
		if len(changed.Movie.Movies) == 0 {
			changed.Movie.Movies = []string{}
			for i = 0; i < len(Moviesprofileitems[email]); i++ {
				changed.Movie.Movies = append(changed.Movie.Movies, Moviesprofileitems[email][i])
			}
		}
		if len(changed.Travel.Flight.Seat) == 0 {
			changed.Travel.Flight.Seat = prev.Travel.Flight.Seat
		}

		TVShowitems := changed.Movie.Tv_shows
		Moviesitems := changed.Movie.Movies

		TVprofileitems[changed.Email] = TVShowitems
		Moviesprofileitems[changed.Email] = Moviesitems
		/* insert new profile in the database */
		DbSaveProfile(changed)
	}

	w.WriteHeader(204)
}

func CreateDb() {
	// Create a new database file and open it
	jb, err := goejdb.Open(dbfile, goejdb.JBOWRITER|goejdb.JBOCREAT|goejdb.JBOTRUNC)
	if err != nil {
		os.Exit(1)
	}
	// Close database
	jb.Close()
}

func DbSaveProfile(box Box) {
	// Open database file
	jb, err := goejdb.Open(dbfile, goejdb.JBOWRITER|goejdb.JBOCREAT)
	if err != nil {
		os.Exit(1)
	}
	// Get or create collection 'profile'
	profile, _ := jb.CreateColl("main", nil)

	// Insert one record:
	// JSON: {'name' : 'Bruce', 'phone' : '333-222-333', 'age' : 58}
	profilerec := map[string]interface{}{
		"Email":          box.Email,
		"Zip":            box.Zip,
		"Country":        box.Country,
		"Profession":     box.Profession,
		"Favorite_color": box.Favourite_color,
		"Travel":         box.Travel,
		"Food":           box.Food,
		"Music":          box.Music,
		"Movie":          box.Movie,
	}

	profilebsrec, _ := bson.Marshal(profilerec)
	profile.SaveBson(profilebsrec)

	fmt.Printf("\nSaved %v", box.Email)
	// Close database
	jb.Close()

	fmt.Printf("\nReplicating %v", box.Email)
	Replicate(profilebsrec)
	fmt.Printf("\nFinished Replicating %v", box.Email)
}

func DbGetProfile(email string) (Box, bool) {
	// Open database file
	jb, err := goejdb.Open(dbfile, goejdb.JBOREADER)
	if err != nil {
		os.Exit(1)
	}

	// Close db, on function return
	defer jb.Close()

	// Get or create collection 'profile'
	profile, _ := jb.CreateColl("main", nil)

	// Now execute query
	// Name starts with 'Bru' string
	res, _ := profile.Find(fmt.Sprintf("{\"Email\" : {\"$begin\" : \"%v\"}}", email))
	fmt.Printf("\nDbGetProfile - Email: %s\n", email)
	fmt.Printf("DbGetProfile - Records found: %d\n", len(res))

	// Now print the result set records
	for i, bs := range res {
		var m map[string]interface{}
		bson.Unmarshal(bs, &m)
		fmt.Println(m)
		var box Box
		box.Email = m["Email"].(string)
		if m["Country"] != nil {
			box.Country = m["Country"].(string)
		}
		if m["Zip"] != nil {
			box.Zip = m["Zip"].(string)
		}
		if m["Profession"] != nil {
			box.Profession = m["Profession"].(string)
		}
		if m["Favourite_color"] != nil {
			box.Favourite_color = m["Favourite_color"].(string)
		}

		if m["Travel"] != nil {
			travel := m["Travel"].(map[string]interface{})
			flight := travel["flight"].(map[string]interface{})
			seat := flight["seat"].(string)
			box.Travel = Travel{Flight{seat}}
		}

		if m["Food"] != nil {
			food := m["Food"].(map[string]interface{})
			typ := food["type"].(string)
			drink := food["drink_alcohol"].(string)
			box.Food.Type = typ
			box.Food.Drink_alcohol = drink
		}

		if m["Music"] != nil {
			food := m["Music"].(map[string]interface{})
			spotify := food["spotify_user_id"].(string)
			box.Music.Spotify_user_id = spotify
		}

		if m["Movie"] != nil {
			movie := m["Movie"].(map[string]interface{})
			films := movie["movies"].([]interface{})
			shows := movie["tv_shows"].([]interface{})
			for _, film := range films {
				box.Movie.Movies = append(box.Movie.Movies, film.(string))
			}
			for _, show := range shows {
				box.Movie.Tv_shows = append(box.Movie.Tv_shows, show.(string))
			}

		}

		// Return the most recent record
		if i == len(res)-1 {
			return box, true
		}
	}
	return Box{}, false
}

func DbSaveMovieProfile(email string, movie Movie) {
	// TODO
}

func DbGetMovieProfile(email string) Movie {
	// TODO
	return Movie{}
}

func DbSaveTvProfile(email string, shows []string) {
	// TODO
}

func DbGetTvProfile(email string) []string {
	// TODO
	return []string{}
}

func Replicate(profilebs []byte) {
	log.Println("\nStarting replication")
	for _, r := range replica {
		log.Println("Replicating profile to:", r)
		SendRPCMsg(r, profilebs)
	}
	log.Println("Finished replication")
}

func SendRPCMsg(server string, profilebs []byte) {
	client, err := rpc.Dial("tcp", server)
	if err != nil {
		log.Print("REPLICATION FAILED:")
		log.Print(err)
		return
	}

	var reply bool
	err = client.Call("Listener.ReceiveProfile", profilebs, &reply)
	if err != nil {
		log.Print("REPLICATION FAILED:")
		log.Print(err)
		return
	}
}

type Listener int

func (l *Listener) ReceiveProfile(profilebs []byte, ack *bool) error {
	var m map[string]interface{}
	bson.Unmarshal(profilebs, &m)
	fmt.Println(m)

	// Open database file
	jb, err := goejdb.Open(dbfile, goejdb.JBOWRITER|goejdb.JBOCREAT)
	if err != nil {
		os.Exit(1)
	}
	// Get or create collection 'profile'
	profile, _ := jb.CreateColl("main", nil)
	profile.SaveBson(profilebs)
	fmt.Println("\nProfile Replication saved")

	// Close database file
	jb.Close()

	return nil
}

func ReceiveRPCMsg() {
	addy, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", rpcServerPort))
	if err != nil {
		log.Fatal(err)
	}

	inbound, err := net.ListenTCP("tcp", addy)
	if err != nil {
		log.Fatal(err)
	}

	listener := new(Listener)
	rpc.Register(listener)
	rpc.Accept(inbound)
}
