package models

import "gopkg.in/mgo.v2/bson"

type Canto struct {
	ID    bson.ObjectId `bson:"_id" json:"id"`
	Book string        `bson:"book" json:"book"`
	Title string        `bson:"title" json:"title"`
	Roman string        `bson:"roman" json:"roman"`
	Arabic  int           `bson:"arabic" json:"arabic"`
	Verse  int           `bson:"verse" json:"verse"`
	Words  int           `bson:"words" json:"words"`
	TextItalian  string        `bson:"textItalian" json:"textItalian"`
	TextEnglish  string        `bson:"textEnglish" json:"textEnglish"`
}

