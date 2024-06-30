package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)
type Block struct{
    Index int // index of block
    Timestamp string //timestamp for data written
    BPM int  //my pulse rate 
    Hash string // hash of present block
    PrevHash string // hash of next block
}
