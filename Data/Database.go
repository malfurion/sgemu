package Data

import (
	"encoding/hex"
	"fmt"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

var (
	Session  *mgo.Session
	CUsers   *mgo.Collection
	CPlayers *mgo.Collection
)

func InitializeDatabase() {
	log.Printf("Connecting to MongoDB...\n")
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(fmt.Sprintf("Connecting to MongoDB has been failed! err:%v\n", err))
	}
	session.SetSyncTimeout(30 * 1000000000)
	err = session.Ping()
	if err != nil {
		panic(fmt.Sprintf("Connecting to MongoDB has been failed! err:%v\n", err))
	}
	log.Println("Connected!")
	Session = session
}

func CreateDatabase() {
	c := Session.DB("SGEmu").C("Users")
	p := Session.DB("SGEmu").C("Players")
	CPlayers = p
	CUsers = c
	index := mgo.Index{
		Key:        []string{"_id", "user", "email"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := CUsers.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	n, _ := CUsers.Find(nil).Count()
	log.Printf("%d Users found!\n", n)

	index = mgo.Index{
		Key:        []string{"_id", "userid"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = CPlayers.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func ClearDatabase() {
	if Session != nil {
		CUsers.RemoveAll(bson.M{})
		CPlayers.RemoveAll(bson.M{})
	}
}

func NewID() string {
	return hex.EncodeToString([]byte(string(bson.NewObjectId())))
}

func NewIID(c *mgo.Collection) uint32 {
	type dummyID struct {
		Seq uint32
	}

	d := dummyID{0}

	change := mgo.Change{Update: bson.M{"$inc": bson.M{"seq": 1}}, ReturnNew: true}
	_, e := c.Find(bson.M{"_id": "users"}).Apply(mgo.Change{Update: change}, &d)
	if e != nil {
		panic(fmt.Sprintf("Could not generate NewIID! collection:%s err:%v\n", c.FullName, e))
	}
	return d.Seq
}

func AddAutoIncrementingField(c *mgo.Collection) {
	i, e := c.Find(bson.M{"_id": "users"}).Count()
	if e != nil {
		panic(fmt.Sprintf("Could not Add Auto Incrementing Field! collection:%s err:%v\n", c.FullName, e))
	}
	if i > 0 {
		return
	}
	c.Insert(bson.M{"_id": "users", "seq": uint32(0)})
}
