package main

import (
    "bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
    "net" // for tcp
	// "net/http" // for http
	"os"
	"time"
    "strconv"

	"github.com/davecgh/go-spew/spew"
	// "github.com/gorilla/mux" //for http
	"github.com/joho/godotenv"
)

//Blockchain 
type Block struct{
    Index int // index of block
    Timestamp string //timestamp for data written
    BPM int  //my pulse rate 
    Hash string // hash of present block
    PrevHash string // hash of next block
}
var Blockchain [] Block // slice of block
var bcServer chan [] Block //handling concurrent blocks

//function to calculte the next hash using the data
func calculateHash(block Block) string{
    record := string(block.Index) + block.Timestamp + string(block.BPM) + block.PrevHash
    h := sha256.New()
    h.Write([]byte(record))
    hashed := h.Sum(nil)
    return hex.EncodeToString(hashed)
}
//function to create a new block
func generateBlock(prevblock Block, BPM int) (Block , error){
    var nextblock Block

    t := time.Now()

    nextblock.Index = prevblock.Index + 1
    nextblock.Timestamp = t.String()
    nextblock.BPM = BPM
    nextblock.PrevHash = prevblock.Hash
    nextblock.Hash = calculateHash(nextblock)

    return nextblock , nil
}
//check if block is valid
func isBlockValid(prevblock , nextblock Block) bool {
    if prevblock.Index+1 != nextblock.Index {
		return false
	}

	if prevblock.Hash != nextblock.PrevHash {
		return false
	}

	if calculateHash(nextblock) != nextblock.Hash {
		return false
	}

	return true
}
//replace the chain
func replaceChain(nextblocks [] Block) {
    if(len(nextblocks) > len(Blockchain)){
        Blockchain = nextblocks
    }
}



//Web Server
/* for http
func run() error{
    mux := makeMuxRouter()
    httpAddr := os.Getenv("ADDR")
    log.Println(httpAddr)
    
    log.Println("Listening on ", os.Getenv("ADDR"))
    s := &http.Server{
        Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
    }

    if err := s.ListenAndServe(); err != nil {
		return err
	}

    return nil
}

func makeMuxRouter() http.Handler{
    muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
    bytes, err := json.MarshalIndent(Blockchain, "", "  ")
    if(err != nil){
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }
    io.WriteString(w, string(bytes))
}

type Message struct{
    BPM int
}
func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
    var m Message

    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&m); err != nil {
		respondWithJSON(w, r, http.StatusBadRequest, r.Body)
		return
	}
    defer r.Body.Close()

    nextblock, err := generateBlock(Blockchain[len(Blockchain)-1], m.BPM)

    if err != nil {
		respondWithJSON(w, r, http.StatusInternalServerError, m)
		return
	}

    if isBlockValid(nextblock, Blockchain[len(Blockchain)-1]) {
		nextblockchain := append(Blockchain, nextblock)
		replaceChain(nextblockchain)
		spew.Dump(Blockchain)
	}

	respondWithJSON(w, r, http.StatusCreated, nextblock)
}

func respondWithJSON(w http.ResponseWriter, r *http.Request, code int, payload interface{}) {
	response, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("HTTP 500: Internal Server Error"))
		return
	}
	w.WriteHeader(code)
	w.Write(response)
}
*/

//Main

func main(){
    err := godotenv.Load()
    if err != nil {
		log.Fatal(err)
	}
    
    bcServer = make(chan [] Block)
//    go func() {
    t := time.Now()
    genesisBlock := Block{0, t.String(), 0, "", ""}
    spew.Dump(genesisBlock)
    Blockchain = append(Blockchain, genesisBlock)
//	}
    server , err := net.Listen("tcp", ":"+os.Getenv("ADDR"))

    if err != nil {
		log.Fatal(err)
	}


    for{
        conn, err := server.Accept()
        
        if err != nil {
            log.Fatal(err)
        }
        go handleConn(conn)
    }
}    
func handleConn(conn net.Conn){
    defer conn.Close()
    io.WriteString(conn , "Enter a new BPM: ")

    scanner := bufio.NewScanner(conn)

    go func(){
        for scanner.Scan(){
            bpm, err := strconv.Atoi(scanner.Text())
            if err != nil {
                log.Printf("%v not a number: %v", scanner.Text(), err)
                continue
            }
            newBlock, err := generateBlock(Blockchain[len(Blockchain)-1], bpm)
            if err != nil {
                log.Println(err)
                continue
            }
            if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
                newBlockchain := append(Blockchain, newBlock)
                replaceChain(newBlockchain)
            }

            bcServer <- Blockchain
            io.WriteString(conn, "\nEnter a new BPM:")
        }
    }()
    go func() {
        for {
            time.Sleep(30 * time.Second)
            output, err := json.Marshal(Blockchain)
            if err != nil {
                log.Fatal(err)
            }
            io.WriteString(conn, string(output))
        }
    }()

    for _ = range bcServer {
        spew.Dump(Blockchain)
    }
}
