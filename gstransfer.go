package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/dgraph-io/badger"
	"github.com/karrick/godirwalk"
)

// server mode
var bind string
var maxSize int64

// client mode
var processID string
var sourceDir string
var dstRule string
var dbPath string
var serverURL string
var gsBucket string

func init() {

	fmt.Println("v0.0.-5")
	fmt.Println("define -bind flag if you want to use server mode")

	// server mode
	flag.StringVar(&bind, "bind", "", "set gstransfer bind ip and port")
	flag.Int64Var(&maxSize, "max-size", 100, "size in MB")

	// client mode
	flag.StringVar(&processID, "id", "", "define process id, this will enable re-upload without duplicate file on app crash")
	flag.StringVar(&sourceDir, "source", "", "full path to your source directory")
	flag.StringVar(&dstRule, "dst-rule", "dir-0/../filename", "dir-0 is root of chosen directory, and so on, and 'filename' is your file original filename, you can add custom prefix")
	flag.StringVar(&dbPath, "db-path", "/tmp/gstransfer", "full path to database folder, will automatically craete if not exist")
	flag.StringVar(&serverURL, "server", "", "your server url")
	flag.StringVar(&gsBucket, "bucket", "", "your bucket name")

	flag.Parse()

}

func main() {

	if bind != "" {
		fmt.Println("binding to", bind)
		serverMode()
	}

	if sourceDir == "" || gsBucket == "" || processID == "" || serverURL == "" {
		panic("required args for client, -source, -bucket, -server, -id")
	}

	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = godirwalk.Walk(sourceDir, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {

			if !de.IsDir() {

				relativePath := strings.Replace(osPathname, sourceDir, "", 1)
				relativePath = strings.Replace(relativePath, "/", "", 1)

				dbKey := []byte(processID + ":" + osPathname)

				_ = db.View(func(txn *badger.Txn) error {
					_, err := txn.Get(dbKey)
					if err != nil {

						object, err := destObjectRewrite(relativePath, dstRule)
						if err == nil {

							log.Printf("[UPLOAD] %s %s\n", de.ModeType(), osPathname)

							extraParams := map[string]string{
								"bucket": gsBucket,
								"object": object,
							}

							log.Println(extraParams)

							resp, err := newfileUploadRequest(serverURL+"/upload", extraParams, "file", osPathname)
							if err != nil {
								log.Println("newfileUploadRequest error", err)
							} else {

								// save current time to divide when request are done
								start := time.Now()

								elapsed := time.Since(start)
								if err != nil {
									log.Println("upload error", err)
								} else {
									log.Printf("uploaded successfully, took %s with resp: %s \n", elapsed, resp)

									// Start a writable transaction.
									txn := db.NewTransaction(true)

									// Use the transaction...
									err := txn.Set(dbKey, []byte("ok"))
									if err != nil {
										log.Println("set err", err)
									}

									// Commit the transaction and check for error.
									if err := txn.Commit(); err != nil {
										log.Println("Commit err", err)
									}

									// Discard ???
									txn.Discard()
								}
							}
						}
					}
					return nil
				})
			}

			return nil
		},
		Unsorted: true, // (optional) set true for faster yet non-deterministic enumeration (see godoc)
	})

}

func serverMode() {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		panic("storage.NewClient: " + err.Error())
	}
	defer client.Close()

	uploadAndTransfer := func(w http.ResponseWriter, r *http.Request) {
		// Parse our multipart form, maxSize << 20 specifies a maximum
		// upload of maxSize MB files.
		r.ParseMultipartForm(maxSize << 20)

		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Error Retrieving the File", err)
			return
		}

		defer file.Close()

		bucket := r.FormValue("bucket")
		object := r.FormValue("object")

		log.Println("incoming object to", bucket, "as", object)

		ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
		defer cancel()

		// Upload an object with storage.Writer.
		wc := client.Bucket(bucket).Object(object).NewWriter(ctx)
		if _, err = io.Copy(wc, file); err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "io.Copy:", err.Error())
			return
		}

		if err := wc.Close(); err != nil {
			w.WriteHeader(500)
			fmt.Fprint(w, "Writer.Close:", err)
			return
		}

		w.Write([]byte("success"))

	}

	http.HandleFunc("/upload", uploadAndTransfer)
	log.Fatal(http.ListenAndServe(bind, nil))
}
