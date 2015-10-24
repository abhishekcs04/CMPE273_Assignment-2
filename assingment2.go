package main

import (
    "fmt"
    "github.com/julienschmidt/httprouter"
    "io/ioutil"
    "encoding/json"
    "net/http"
    "strings"
    "gopkg.in/mgo.v2/bson"
    "gopkg.in/mgo.v2"
)






type Reply struct
{
Id     bson.ObjectId `json:"id" bson:"_id"`
Name string	`json:"name" bson:"name"`
Address string	`json:"address" bson:"address" `
City string		`json:"city"  bson:"city"`
State string	`json:"state"  bson:"state"`
ZipCode string	`json:"zip"  bson:"zip" `
Coordinate struct 
{
Lat float64 `json:"lat"   bson:"lat"`
Lng float64 `json:"lng"   bson:"lng"`
} `json:"coordinate" bson:"coordinate"`
}


type GCoordinates struct {
	Results []struct {
		AddressComponents []struct {
			LongName string `json:"long_name"`
			ShortName string `json:"short_name"`
			Types []string `json:"types"`
		} `json:"address_components"`
		FormattedAddress string `json:"formatted_address"`
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
			LocationType string `json:"location_type"`
			Viewport struct {
				Northeast struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"northeast"`
				Southwest struct {
					Lat float64 `json:"lat"`
					Lng float64 `json:"lng"`
				} `json:"southwest"`
			} `json:"viewport"`
		} `json:"geometry"`
		PlaceID string `json:"place_id"`
		Types []string `json:"types"`
	} `json:"results"`
	Status string `json:"status"`
}

func getSession() *mgo.Session {  
    session, err := mgo.Dial("mongodb://user:pass@ds047901.mongolab.com:47901/cmpe273")
    if err != nil {
        panic(err)
    }
    return session
}

type UserController struct {  
    session *mgo.Session
}
   
func NewUserController(s *mgo.Session) *UserController {  
    return &UserController{s}
}

func createLoc (rw http.ResponseWriter, r *http.Request, _ httprouter.Params) {  
    
uc := NewUserController(getSession()) 

   var C GCoordinates
    var V Reply
    
    json.NewDecoder(r.Body).Decode(&V)   
  UrlStr:=  V.Address+","+V.City+","+V.State+","+V.ZipCode
  UrlStr = strings.Replace(UrlStr," ","+",-1)
  UrlStr = "https://maps.google.com/maps/api/geocode/json?address="+UrlStr+"&sensor=false"
    
client := &http.Client{}
req, _ := http.NewRequest("GET", UrlStr, nil)
resp,_:= client.Do(req)

if( resp.StatusCode >= 200 && resp.StatusCode < 300 ) {
  body, _ := ioutil.ReadAll(resp.Body) 
   _= json.Unmarshal(body, &C)
     }
     
    for _,Sample := range C.Results {
    	V.Coordinate.Lat= Sample.Geometry.Location.Lat
	V.Coordinate.Lng = Sample.Geometry.Location.Lng
         }
             V.Id = bson.NewObjectId()
    uc.session.DB("cmpe273").C("users").Insert(V)
   UJ, _ := json.Marshal(V)
     fmt.Fprintf(rw, "%s", UJ)
    	  }
	  
	  
func getLoc (w http.ResponseWriter, r *http.Request, p httprouter.Params) {  
    
uc := NewUserController(getSession()) 
id:= p.ByName("id")

if !bson.IsObjectIdHex(id) {
        w.WriteHeader(404)
        return
    }
    
    oid := bson.ObjectIdHex(id)
    
   v:= Reply{}
    
 if err := uc.session.DB("cmpe273").C("users").FindId(oid).One(&v); err != nil {
        w.WriteHeader(404)
        return
    }

    uj, _ := json.Marshal(v)
	
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    fmt.Fprintf(w, "%s", uj)
    
    
    	  }
	  
	  
func removeLoc (w http.ResponseWriter, r *http.Request, p httprouter.Params) {  

uc := NewUserController(getSession()) 

    id := p.ByName("id")

    if !bson.IsObjectIdHex(id) {
        w.WriteHeader(404)
        return
    }

    oid := bson.ObjectIdHex(id)

    if err := uc.session.DB("cmpe273").C("users").RemoveId(oid); err != nil {
        w.WriteHeader(404)
        return
    }

    w.WriteHeader(200)
}	  

 func updateloc (rw http.ResponseWriter, r *http.Request, p httprouter.Params) {  
    
uc := NewUserController(getSession()) 

   var C GCoordinates
    var V Reply
	id:= p.ByName("id")

	if !bson.IsObjectIdHex(id) {
        rw.WriteHeader(404)
        return
    }
    
    oid := bson.ObjectIdHex(id)
    
    json.NewDecoder(r.Body).Decode(&V)   
  UrlStr:=  V.Address+","+V.City+","+V.State+","+V.ZipCode
  UrlStr = strings.Replace(UrlStr," ","+",-1)
  UrlStr = "https://maps.google.com/maps/api/geocode/json?address="+UrlStr+"&sensor=false"
    
client := &http.Client{}
req, _ := http.NewRequest("GET", UrlStr, nil)
resp,_:= client.Do(req)

if( resp.StatusCode >= 200 && resp.StatusCode < 300 ) {
  body, _ := ioutil.ReadAll(resp.Body) 
   _= json.Unmarshal(body, &C)
     }
     
    for _,Sample := range C.Results {
    	V.Coordinate.Lat= Sample.Geometry.Location.Lat
	V.Coordinate.Lng = Sample.Geometry.Location.Lng
         }
            


    uc.session.DB("cmpe273").C("users").Update(bson.M{ "_id":oid},bson.M{"$set":bson.M{"address":V.Address,"city":V.City,"state":V.State,"zip":V.ZipCode,"coordinate.lat":V.Coordinate.Lat,"coordinate.lng":V.Coordinate.Lng}})
    if err := uc.session.DB("cmpe273").C("users").FindId(oid).One(&V); err != nil {
        rw.WriteHeader(404)
        return
    }
   
    uj, _ := json.Marshal(V)

    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", uj)
    	  }
	  	  
	  
func main() {

  mux := httprouter.New()
  
    mux.GET("/locations/:id", getLoc)
    
    mux.POST("/locations",createLoc)
	
	mux.PUT("/locations/:id",updateloc)
    
    mux.DELETE("/locations/:id", removeLoc)
    
    server := http.Server{
            Addr:        "0.0.0.0:8080",
            Handler: mux,
    }
    server.ListenAndServe()
}