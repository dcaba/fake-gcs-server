// Copyright 2017 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fakestorage

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

// CreateBucket creates a bucket inside the server, so any API calls that
// require the bucket name will recognize this bucket.
//
// If the bucket already exists, this method does nothing but panics if props differs
func (s *Server) CreateBucket(name string, versioningEnabled bool) {
	err := s.backend.CreateBucket(name, versioningEnabled)
	if err != nil {
		panic(err)
	}
}

// createBucketByPost handles a POST request to create a bucket
func (s *Server) createBucketByPost(w http.ResponseWriter, r *http.Request) {
	// Minimal version of Bucket from google.golang.org/api/storage/v1

	var data struct {
		Name       string            `json:"name,omitempty"`
		Versioning *bucketVersioning `json:"versioning,omitempty"`
	}

	// Read the bucket props from the request body JSON
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	name := data.Name
	versioning := false
	if data.Versioning != nil {
		versioning = data.Versioning.Enabled
	}
	// Create the named bucket
	if err := s.backend.CreateBucket(name, versioning); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created bucket:
	resp := newBucketResponse(name, versioning)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) listBuckets(w http.ResponseWriter, r *http.Request) {
	buckets, err := s.backend.ListBuckets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := newListBucketsResponse(buckets)
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) getBucket(w http.ResponseWriter, r *http.Request) {
	bucketName := mux.Vars(r)["bucketName"]
	encoder := json.NewEncoder(w)
	bucket, err := s.backend.GetBucket(bucketName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		err := newErrorResponse(http.StatusNotFound, "Not found", nil)
		encoder.Encode(err)
		return
	}
	resp := newBucketResponse(bucket.Name, bucket.VersioningEnabled)
	w.WriteHeader(http.StatusOK)
	encoder.Encode(resp)
}
