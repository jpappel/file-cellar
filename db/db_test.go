package db

import (
	"context"
	"database/sql"
	"testing"
)

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
			t.Logf("Incorrect resolved url, expected %s but got %s\n", expected, url)
			t.Fail()
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
