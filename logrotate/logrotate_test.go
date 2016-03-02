package logrotate

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func Test_gzipLogFile(t *testing.T) {
	var (
		logFileContent = "foo bar foo bar foo bar foo bar\n foo bar foo bar foo bar foo bar foo bar"
	)

	logFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		logFile.Close()
		os.Remove(logFile.Name())
	}()

	if err := ioutil.WriteFile(logFile.Name(), []byte(logFileContent), 0600); err != nil {
		t.Fatal(err)
	}

	r := NewRotator(logFile.Name())
	archive, err := r.gzipLogFile()
	if err != nil {
		t.Fatal(err)
	}

	//logFile must be empty
	cont, err := ioutil.ReadAll(logFile)
	if string(cont) != "" {
		t.Fatalf("log file not empty: %s", string(cont))
	}

	//archive must contains logFileContent
	archiveFile, err := os.Open(archive)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		archiveFile.Close()
		os.Remove(archiveFile.Name())
	}()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		t.Fatal(err)
	}
	defer gzipReader.Close()

	archiveContent, err := ioutil.ReadAll(gzipReader)
	if err != nil {
		t.Fatal(err)
	}
	if string(archiveContent) != logFileContent {
		t.Fatalf("archive content doesn't match original content: %s", string(archiveContent))
	}
}

func Test_cleanupOldArchives(t *testing.T) {
	nbArchives := 10

	r := NewRotator(path.Join(os.TempDir(), "logfile.log"))

	archives := []string{}
	for i := 0; i < nbArchives; i++ {
		archives = append(archives, fmt.Sprintf("%s/logfile.log_2006010215040%d.gz", os.TempDir(), i))
	}

	//create archive files
	for _, a := range archives {
		if _, err := os.Create(a); err != nil {
			t.Fatal(err)
		}
		defer os.Remove(a)
	}

	foundArchives, err := r.relatedGZFiles()
	if err != nil {
		t.Fatal(err)
	}
	if len(foundArchives) != nbArchives {
		t.Fatalf("excepted to find %d archives, found %d", nbArchives, len(foundArchives))
	}

	//cleanup
	if err := r.cleanupOldArchives(); err != nil {
		t.Fatal(err)
	}

	//should only be
	foundArchives, err = r.relatedGZFiles()
	if len(foundArchives) != r.ArchiveRetainCount {
		t.Fatalf("expected to find %d archives, got %d", r.ArchiveRetainCount, len(foundArchives))
	}
}

func Test_StartWatching(t *testing.T) {
	logFile, err := ioutil.TempFile("", "psdock_")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		logFile.Close()
		os.Remove(logFile.Name())
	}()

	r := NewRotator(logFile.Name())
	r.RotationDelay = 230 * time.Millisecond
	r.ArchiveRetainCount = 1

	//create a routine that continuously write on the file
	go func() {
		for {
			if _, err := logFile.Write([]byte("foo bar\n")); err != nil {
				t.Fatal(err)
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	go r.StartWatching()

	time.Sleep(1 * time.Second)

	r.StopWatching()

	//only one archive should be left
	foundArchives, err := r.relatedGZFiles()
	if err != nil {
		t.Fatal(err)
	}

	if len(foundArchives) != 1 {
		t.Fatalf("expected one archive, got %d", len(foundArchives))
	}
	os.Remove(foundArchives[0])
}
