package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/jcelliott/lumber"
)

const Version = "1.0.0"

type User struct {
	Name    string
	Age     json.Number
	Contact string
	Address Address
}

type Address struct {
	City    string
	State   string
	Country string
	Pincode json.Number
}

type (
	Driver struct {
		Dir     string
		Mutex   sync.Mutex
		Mutexes map[string]*sync.Mutex
		Log     Logger
	}

	Logger interface {
		Fatal(string, ...interface{})
		Error(string, ...interface{})
		Warn(string, ...interface{})
		Info(string, ...interface{})
		Debug(string, ...interface{})
		Trace(string, ...interface{})
	}
)

type Options struct {
	Logger
}

func main() {
	dir := "./database" // Changed the directory to a subdirectory called 'database'

	db, err := New(dir, nil)
	if err != nil {
		fmt.Println("Error initializing the DB:", err)
	}

	Employees := []User{
		{"Rohan", "18", "3257459824", Address{"Bhopal", "Madhya pradesh", "India", "462022"}},
		{"Adarsh", "20", "27478732732", Address{"Bhopal", "Madhya pradesh", "India", "462022"}},
		{"Aakriti", "21", "9054754126", Address{"Bhopal", "Madhya pradesh", "India", "462022"}},
		{"Pritam", "16", "7895412354", Address{"Bhopal", "Madhya pradesh", "India", "462022"}},
		{"Harshit", "20", "7845963254", Address{"Bhopal", "Madhya pradesh", "India", "462022"}},
	}

	for _, val := range Employees {
		if err := db.Write("users", val.Name, val); err != nil { // Changed 'db.Write' to 'db.Write' with error handling
			fmt.Println("Error writing to database:", err)
		}
	}

	records, err := db.ReadAll("users")
	if err != nil {
		fmt.Println("Error reading all the users in the database:", err)
	}
	fmt.Println(records)

	allUsers := []User{}

	for _, f := range records {
		employeeFound := User{}
		if err := json.Unmarshal([]byte(f), &employeeFound); err != nil {
			fmt.Println("Error unmarshaling the records:", err)
		}
		allUsers = append(allUsers, employeeFound)
	}
	fmt.Println(allUsers)
}

func New(dir string, options *Options) (*Driver, error) {
	dir = filepath.Clean(dir)

	opts := Options{}

	if options != nil {
		opts = *options
	}

	if opts.Logger == nil {
		opts.Logger = lumber.NewConsoleLogger(lumber.INFO)
	}

	driver := Driver{
		Dir:     dir,
		Mutexes: make(map[string]*sync.Mutex),
		Log:     opts.Logger,
	}

	if _, err := os.Stat(dir); err == nil {
		opts.Logger.Debug("Using '%s' (Database already exists)", dir)
		return &driver, nil
	}

	opts.Logger.Debug("Creating a Database at '%s' ...", dir)
	if err := os.Mkdir(dir, 0755); err != nil {
		return nil, err // Handle the error properly
	}
	return &driver, nil
}

func Stat(path string) (fi os.FileInfo, err error) {
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}
	return
}

func (d *Driver) Write(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missing Collection - no place to save record!")
	}

	if resource == "" {
		return fmt.Errorf("Missing Resource - unable to save record(no name)!")
	}

	mutex := d.GetOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	dir := filepath.Join(d.Dir, collection)
	if err := os.MkdirAll(dir, 0755); err != nil { // Use MkdirAll to create directory and all parent directories if they don't exist
		return err
	}

	finalPath := filepath.Join(dir, resource+".json")
	tempPath := finalPath + ".tmp"

	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	b = append(b, byte('\n'))

	if err := ioutil.WriteFile(tempPath, b, 0644); err != nil {
		return err
	}

	return os.Rename(tempPath, finalPath)
}

func (d *Driver) Read(collection, resource string, v interface{}) error {
	if collection == "" {
		return fmt.Errorf("Missing collection - Unable to read!")
	}

	if resource == "" {
		return fmt.Errorf("Missing resource - unable to read record(no name)! ")
	}

	record := filepath.Join(d.Dir, collection, resource)
	if _, err := Stat(record); err != nil {
		return err
	}

	b, err := ioutil.ReadFile(record + ".json")
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

func (d *Driver) ReadAll(collection string) ([]string, error) {
	if collection == "" {
		return nil, fmt.Errorf("Missing collection - Unable to read!")
	}

	dir := filepath.Join(d.Dir, collection)
	if _, err := Stat(dir); err != nil {
		return nil, err
	}

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var records []string
	for _, file := range files {
		b, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return nil, err
		}

		records = append(records, string(b))
	}

	return records, nil
}

func (d *Driver) Delete(collection, resource string) error {
	path := filepath.Join(d.Dir, collection, resource)
	mutex := d.GetOrCreateMutex(collection)
	mutex.Lock()
	defer mutex.Unlock()

	fi, err := Stat(path)
	if err != nil {
		return fmt.Errorf("unable to find file or directory named %v\n", path)
	}

	if fi.Mode().IsDir() {
		return os.RemoveAll(path)
	} else if fi.Mode().IsRegular() {
		return os.Remove(path + ".json")
	}

	return nil
}

func (d *Driver) GetOrCreateMutex(collection string) *sync.Mutex {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	m, ok := d.Mutexes[collection]
	if !ok {
		m = &sync.Mutex{}
		d.Mutexes[collection] = m
	}

	return m
}
