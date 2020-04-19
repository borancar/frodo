package frodo_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/agnivade/frodo"
)

func TestRead(t *testing.T) {
	err := frodo.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer frodo.Cleanup()

	go func() {
		for err := range frodo.Err() {
			t.Error(err)
		}
	}()

	var wg sync.WaitGroup

	helper := func(file string) {
		wg.Add(1)
		expected, err := ioutil.ReadFile(file)
		if err != nil {
			t.Error(err)
			return
		}
		err = frodo.ReadFile(file, func(buf []byte) {
			defer wg.Done()
			if !bytes.Equal(buf, expected) {
				t.Errorf("unexpected result. Got %v, expected %v", buf, expected)
			}
		})
		if err != nil {
			t.Error(err)
		}
		frodo.Poll()
		wg.Wait()
	}

	t.Run("ZeroByte", func(t *testing.T) {
		helper("testdata/zero_byte.txt")
	})

	t.Run("TwoBytes", func(t *testing.T) {
		helper("testdata/two_bytes.txt")
	})

	t.Run("MediumFile", func(t *testing.T) {
		helper("testdata/ssa.html")
	})

	t.Run("LargeFile", func(t *testing.T) {
		helper("testdata/coverage.out")
	})
}

func TestQueueThreshold(t *testing.T) {
	err := frodo.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer frodo.Cleanup()

	go func() {
		for err := range frodo.Err() {
			t.Error(err)
		}
	}()
	expected, err := ioutil.ReadFile("testdata/ssa.html")
	if err != nil {
		t.Error(err)
		return
	}

	var wg sync.WaitGroup
	wg.Add(6)

	// Trigger 6 reads and verify that results come,
	// without needing to call Poll.
	for i := 0; i < 6; i++ {
		err = frodo.ReadFile("testdata/ssa.html", func(buf []byte) {
			defer wg.Done()
			if !bytes.Equal(buf, expected) {
				t.Errorf("unexpected result. Got %v, expected %v", buf, expected)
			}
		})
	}
	wg.Wait()
}

func TestWrite(t *testing.T) {
	err := frodo.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer frodo.Cleanup()

	go func() {
		for err := range frodo.Err() {
			t.Error(err)
		}
	}()

	dir, err := ioutil.TempDir("", "frodo")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	var wg sync.WaitGroup

	helper := func(file string) {
		wg.Add(1)
		input, err := ioutil.ReadFile("testdata/" + file)
		if err != nil {
			t.Error(err)
			return
		}

		err = frodo.WriteFile(filepath.Join(dir, file), input, 0644, func(n int) {
			defer wg.Done()
			if n != len(input) {
				t.Errorf("unexpected result. Got %d, expected %d bytes", n, len(input))
			}
		})
		if err != nil {
			t.Error(err)
		}
		frodo.Poll()
		wg.Wait()
		got, err := ioutil.ReadFile(filepath.Join(dir, file))
		if err != nil {
			t.Error(err)
			return
		}
		if !bytes.Equal(got, input) {
			t.Errorf("unexpected result. Got %v, expected %v", got, input)
		}
	}

	t.Run("ZeroByte", func(t *testing.T) {
		helper("zero_byte.txt")
	})

	t.Run("MediumFile", func(t *testing.T) {
		helper("ssa.html")
	})

	t.Run("LargeFile", func(t *testing.T) {
		helper("coverage.out")
	})
}

var globalBuf []byte

func BenchmarkRead(b *testing.B) {
	b.Run("stdlib", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buf, err := ioutil.ReadFile("testdata/zero_byte.txt")
			if err != nil {
				b.Error(err)
			}
			buf, err = ioutil.ReadFile("testdata/ssa.html")
			if err != nil {
				b.Error(err)
			}
			buf, err = ioutil.ReadFile("testdata/coverage.out")
			if err != nil {
				b.Error(err)
			}
			globalBuf = buf
		}
	})

	b.Run("stdlib", func(b *testing.B) {
		frodo.Init()
		defer frodo.Cleanup()
		go func() {
			for err := range frodo.Err() {
				b.Error(err)
			}
		}()
		for i := 0; i < b.N; i++ {
			err := frodo.ReadFile("testdata/zero_byte.txt", func(buf []byte) {
				globalBuf = buf
			})
			if err != nil {
				b.Error(err)
			}
			err = frodo.ReadFile("testdata/ssa.html", func(buf []byte) {
				globalBuf = buf
			})
			if err != nil {
				b.Error(err)
			}
			err = frodo.ReadFile("testdata/coverage.out", func(buf []byte) {
				globalBuf = buf
			})
			if err != nil {
				b.Error(err)
			}
		}
	})
}
