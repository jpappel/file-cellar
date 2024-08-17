package db

import (
	"context"
	"database/sql"
	"file-cellar/storage"
	"testing"
	"time"
)

func printMismatch[T any](p func(string, ...any), name string, expected T, recieved T) {
	p("Incorrect %s, expected %v != %v\n", name, expected, recieved)
}

func newTestManager() (*Manager, error) {
	m, err := GetManager(":memory:", SQLITE_DEFAULT_PRAGMAS)
	if err != nil {
		return nil, err
	}

	err = InitTables(m.db)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func sampleData(m *Manager) error {
	db := m.db

	_, err := db.Exec(`
    INSERT INTO drivers(name)
    VALUES
    ('local'),
    ('network')
    `)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
    INSERT INTO bins(driverID, name, externalURL, internalURL, redirect)
    VALUES
    (1, 'slow hard drive', 'media', '/mount/slow', false),
    (1, 'fast ssd', 'games', '/mount/zyoom', false),
    (2, 'home NAS', 'homelab/nas', 'https://myhomenas.local', true)
    `)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
    INSERT INTO files(binID, name, hash, size, relPath, uploadTimestamp)
    VALUES
    (1, 'sentimental video', 'af8182a217f6c4ae4abb6d52951f6e7a2cac3a4d59889e4a7a3cce87ac0ae508', 6e8, 'oldvid.mp4', 1000209017),
    (1, 'marriage photo', 'a0856e75fc1f1ec0d2fed17d534fbc1756770dbb0cc83788cbf8ca861c885fc0', 3.072e4, 'WeddingAltar5.jpg', 451309817),
    (2, 'dota2', '15c11ed3bd0eb92d6d54de44b36131643268e28f4aac9229f83231a0670c290c', 55e9, 'Dota2Beta', 1373370617),
    (3, 'I saw the tv glow', '7b1a56dfcba8ce808cb6392e2403f895afb1f210b85b7d3ad324d365432f01fa', 1.9e9 ,'I_Saw_The_TV_Glow_2024.mp4', 1718538617)
    `)

	localDriver := new(storage.LocalDriver)
	localDriver.SetId(1)
	localDriver.SetName("local")
	m.Drivers["local"] = localDriver

	// TODO: change to correct driver type
	networkDriver := new(storage.LocalDriver)
	networkDriver.SetId(2)
	networkDriver.SetName("network")
	m.Drivers["network"] = networkDriver

	bin := new(storage.Bin)
	bin.Id = 1
	bin.Name = "slow hard drive"
	bin.Path.External = "media"
	bin.Path.Internal = "/mount/slow"
	bin.Redirect = false
	bin.Driver = localDriver
	m.Bins[bin.Name] = bin

	bin = new(storage.Bin)
	bin.Id = 2
	bin.Name = "fast ssd"
	bin.Path.External = "games"
	bin.Path.Internal = "/mount/zyoom"
	bin.Redirect = false
	bin.Driver = localDriver
	m.Bins["fast ssd"] = bin

	bin = new(storage.Bin)
	bin.Id = 3
	bin.Name = "home NAS"
	bin.Path.External = "homelab/nas"
	bin.Path.Internal = "https://myhomenas.local"
	bin.Redirect = true
	bin.Driver = networkDriver
	m.Bins["home NAS"] = bin

	return err
}

func TestResolve(t *testing.T) {
	m, err := newTestManager()
	if err != nil {
		t.Logf("Error creating manager for testing: %v\n", err)
		t.FailNow()
	}
	defer m.Close()

	err = sampleData(m)
	if err != nil {
		t.Logf("Error adding sample data to manager for testing: %v\n", err)
		t.FailNow()
	}

	testGoodCase := func(expected string, uri string) {
		url, err := m.Resolve(context.Background(), uri)
		if err != nil {
			t.Log(err)
			t.FailNow()
		}

		if url != expected {
			printMismatch(t.Errorf, "resolved url", expected, url)
		}

	}

	testNoMatchCase := func(uri string) {
		url, err := m.Resolve(context.Background(), uri)
		if err != sql.ErrNoRows {
			t.Logf("Incorrect error type, expected %v but got %s\n", sql.ErrNoRows, err)
			t.Logf("Resolved Url to %s", url)
			t.Fail()
		}
	}

	t.Log("Testing Resolveable Cases")
	testGoodCase("https://myhomenas.local/I_Saw_The_TV_Glow_2024.mp4", "I_Saw_The_TV_Glow_2024.mp4")
	testGoodCase("/mount/slow/oldvid.mp4", "oldvid.mp4")
	testGoodCase("/mount/slow/WeddingAltar5.jpg", "WeddingAltar5.jpg")
	testGoodCase("/mount/zyoom/Dota2Beta", "Dota2Beta")

	t.Log("Testing Unresolveable Cases")
	testNoMatchCase("bar")
	testNoMatchCase("passwords.txt")
	testNoMatchCase("env")
	testNoMatchCase("pirated_media.fbi")
}

func TestGetFile(t *testing.T) {
	m, err := newTestManager()
	if err != nil {
		t.Logf("Error creating manager for testing: %v\n", err)
		t.FailNow()
	}
	defer m.Close()

	err = sampleData(m)
	if err != nil {
		t.Logf("Error adding sample data to manager for testing: %v\n", err)
		t.FailNow()
	}

	ctx := context.Background()

	testGoodCase := func(expected storage.File, uri string) {
		f, err := m.GetFile(ctx, uri)
		if err != nil {
			t.Errorf("Failed to get file %s: %v", uri, err)
			return
		}
		if !expected.Equal(*f) {
			t.Log("Incorrect file:")
			if expected.Name != f.Name {
				printMismatch(t.Logf, "name", expected.Name, f.Name)
			}
			if expected.Hash != f.Hash {
				printMismatch(t.Logf, "hash", expected.Hash, f.Hash)
			}
			if expected.Size != f.Size {
				printMismatch(t.Logf, "size", expected.Size, f.Size)
			}
			if expected.RelPath != f.RelPath {
				printMismatch(t.Logf, "relPath", expected.RelPath, f.RelPath)
			}
			if expected.Bin != f.Bin {
				printMismatch(t.Logf, "Bin ptr", expected.Bin, f.Bin)
			}
			if !expected.UploadTimestamp.Equal(f.UploadTimestamp) {
				printMismatch(t.Logf, "time", expected.UploadTimestamp, f.UploadTimestamp)
			}

			t.Fail()
			return
		}
	}

	testBadCase := func(expected error, uri string) {
		f, err := m.GetFile(ctx, uri)
		if err != expected {
			printMismatch(t.Logf, "error type", expected, err)
			t.Logf("Got file %s\n", f)
			t.Fail()
			return
		}
	}

	t.Log("Testing Existing Files")
	expected := storage.File{
		Name:            "sentimental video",
		Hash:            "af8182a217f6c4ae4abb6d52951f6e7a2cac3a4d59889e4a7a3cce87ac0ae508",
		Size:            6e8,
		RelPath:         "oldvid.mp4",
		Bin:             m.Bins["slow hard drive"],
		UploadTimestamp: time.Unix(1000209017, 0),
	}
	testGoodCase(expected, "oldvid.mp4")

	expected = storage.File{
		Name:            "marriage photo",
		Hash:            "a0856e75fc1f1ec0d2fed17d534fbc1756770dbb0cc83788cbf8ca861c885fc0",
		Size:            3.072e4,
		RelPath:         "WeddingAltar5.jpg",
		UploadTimestamp: time.Unix(451309817, 0),
		Bin:             m.Bins["slow hard drive"],
	}
	testGoodCase(expected, "WeddingAltar5.jpg")

	expected = storage.File{
		Name:            "dota2",
		Hash:            "15c11ed3bd0eb92d6d54de44b36131643268e28f4aac9229f83231a0670c290c",
		Size:            55e9,
		RelPath:         "Dota2Beta",
		UploadTimestamp: time.Unix(1373370617, 0),
		Bin:             m.Bins["fast ssd"],
	}
	testGoodCase(expected, "Dota2Beta")

	expected = storage.File{
		Name:            "I saw the tv glow",
		Hash:            "7b1a56dfcba8ce808cb6392e2403f895afb1f210b85b7d3ad324d365432f01fa",
		Size:            1.9e9,
		RelPath:         "I_Saw_The_TV_Glow_2024.mp4",
		UploadTimestamp: time.Unix(1718538617, 0),
		Bin:             m.Bins["home NAS"],
	}
	testGoodCase(expected, "I_Saw_The_TV_Glow_2024.mp4")

	t.Log("Testing Non-Existing Files")
	testBadCase(sql.ErrNoRows, "bingbong")
}
