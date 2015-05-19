package main

import (
	"bytes"
	"encoding/gob"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
	"log"
	"sort"
	"strings"
	"time"
)

// HA HA, joke's on you! ENTIRE DB IS FILE!
const filename = "sparkledb"
const bucketName = "mister-sparkleo"

type SparkleDatabase struct {
	Sparkles []Sparkle
}

func (sparkledb *SparkleDatabase) Save() {
	// Persist the database to file
	var data bytes.Buffer
	contents := gob.NewEncoder(&data)
	err := contents.Encode(sparkledb)
	if err != nil {
		panic(err)
	}

	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	// Open Bucket
	s := s3.New(auth, aws.USEast)

	// Load the database from an S3 bucket
	bucket := s.Bucket(bucketName)

	err = bucket.Put(filename, data.Bytes(), "text/plain", s3.BucketOwnerFull)
	if err != nil {
		panic(err.Error())
	}
}

func LoadDB() SparkleDatabase {
	// The AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY environment variables are used.
	auth, err := aws.EnvAuth()
	if err != nil {
		panic(err.Error())
	}

	// Open Bucket
	s := s3.New(auth, aws.USEast)

	// Load the database from an S3 bucket
	bucket := s.Bucket(bucketName)

	// Create a bytes.Buffer
	n, err := bucket.Get(filename)
	if err != nil {
		panic(err)
	}

	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)

	var sparkleDB SparkleDatabase
	err = dec.Decode(&sparkleDB)

	if err != nil {
		log.Print("There was an error loading the sparkle database. Using a blank one.")
	}

	return sparkleDB
}

func (sparkledb *SparkleDatabase) AddSparkle(sparkle Sparkle) Leader {
	// Add a sparkle to the database
	sparkle.Time = time.Now()
	sparkledb.Sparkles = append(sparkledb.Sparkles, sparkle)

	// After the sparkle has been added, save the data file
	sparkledb.Save()

	// Return the receiver record so that Hubot can report the users total sparkles
	receivers := sparkledb.Receivers()
	var recipient Leader
	for _, v := range receivers {
		if v.Name == sparkle.Sparklee {
			recipient = v
		}
	}

	return recipient
}

func (s *SparkleDatabase) Givers() []Leader {
	var g = make(map[string]int)
	for _, v := range s.Sparkles {
		g[v.Sparkler]++
	}

	var leaders []Leader
	for k, v := range g {
		leader := Leader{Name: k, Score: v}
		leaders = append(leaders, leader)
	}

	return leaders
}

func (s *SparkleDatabase) Receivers() []Leader {
	var g = make(map[string]int)
	for _, v := range s.Sparkles {
		g[v.Sparklee]++
	}

	var leaders []Leader
	for k, v := range g {
		leader := Leader{Name: k, Score: v}
		leaders = append(leaders, leader)
	}

	return leaders
}

type ByScore []Leader

func (a ByScore) Len() int           { return len(a) }
func (a ByScore) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByScore) Less(i, j int) bool { return a[i].Score < a[j].Score }

func (sparkledb *SparkleDatabase) TopGiven() []Leader {
	leaders := sparkledb.Givers()
	sort.Sort(sort.Reverse(ByScore(leaders)))
	return leaders
}

func (sparkledb *SparkleDatabase) TopReceived() []Leader {
	leaders := sparkledb.Receivers()
	sort.Sort(sort.Reverse(ByScore(leaders)))
	return leaders
}

func (db *SparkleDatabase) SparklesForUser(user string) []Sparkle {
	// Return all the sparkles for <user>
	var list []Sparkle
	for _, v := range db.Sparkles {
		if strings.ToLower(v.Sparklee) == strings.ToLower(user) {
			list = append(list, v)
		}
	}

	return list
}
