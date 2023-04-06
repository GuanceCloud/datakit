This project is forked from [BurntSushi/toml](https://github.com/BurntSushi/toml) for comment-preserving toml parser and generator, which especially suit for machine-edit config file.

This library requires Go 1.13 or newer; add it to your go.mod with:

    $ go get github.com/GuanceCloud/toml


### Examples

The input.toml file:

```toml
# preserving comment example
# bind rules:
# 1. document comments will be bound to below nearest key
# 2. comment at the end of line will be bound to prefix key

version = "v1.0.1" # bound to version key

# below list release history
release_history = [
    {date = "2023-03-01", version = "v0.3", features = [
        "1. [SESSION] Supports Sysmon for easy monitoring of system resources usage, such as CPU, memory, network",
        "2. Start WindTerm and select the profiles directory and quit.", # comment for features[2]
        "3. [TAB] Restore the last modified tab name. #626" # comments for features[3]
    ]}, # comment for release_history[0]

    # this comment will be bound to below reeases[1]
    {date = "2023-02-01", version = "v0.2"}, # this comment will also be bound to releases[1]
    {date = "2023-01-01", version = "v0.1"}, # comment for release[2]
]

[author]
name = "zhangsan"  # name
age = +32  # age
gender = "male" # gender
"marital status" = true # marital status
# document comment for hobbies
hobbies = [ #comment for hobbies
    [
        "basketball", # 篮球
        "football" # 足球
    ],
    ["tennis"], # 网球
    # this is full line comment for sub array
    [
        "Snooker", # sub array[0]
        "Billy Billiards", # sub array[1]
        "Pyramid" # sub array[2]
    ], # this is line tail comment for sub array
]    # also comment for hobbies

# the comment for first country
[[country]]
name = "China" # comment bound to China
code = "CN"  # comment bound to CN
continent = "Asia" # comment bound to Asia
# main city list
main_cities = [
    "Beijing", # bounds to main_cities[0]
    "Shanghai", # bounds to main_cities[1]
    "Guangzhou" # bounds to main_cities[2]
] # main city list of China over

# the second country: America
# The United States of America (U.S.A. or USA), commonly known as the United States (U.S. or US) or America,
# is a country primarily located in North America. It consists of 50 states, a federal district,
# five major unincorporated territories, nine Minor Outlying Islands,[h] and 326 Indian reservations.
[[country]]
# comment for America
name = "America"
code = "USA"
continent = "North America"

# the largest cities in America
# including New York, Los Angeles ...
#
main_cities = [
    "New York", # bounds to main_cities[0]: New York
    "Los Angeles", # bounds to main_cities[1]: Los Angeles
    "Chicago", # Chicago is an international hub for finance, culture, commerce, industry, education, technology, telecommunications, and transportation
    "San Francisco"
] # main city list of America over

# this comment will be bound to country[1].state[0]
[[country.state]]
name = "Massachusetts"
area = 21000
population = 7029917
latitude = [
    "41°14′ N", # latitude from
    "42°53′ N" # latitude to
]
# comment for Longitude
longitude = [
    "69°56′ W", # Longitude from
    "73°30′ W" # Longitude to
]

[[country.state]] # this is comment for country[1].state[1]
name = "Texas"
capital = "Austin"
area = 696241
population = 29145505
latitude = ["25°50′ N", "36°30′ N"]
longitude = ["93°31′ W", "106°39′ W"]



[[country]]
name = "Germany"
code = "DE"
continent = "Europe"
main_cities = [
    "Hamburg",
    "Munich",
    "Frankfurt"
] # comment bound to main_cities

# this and below comments will be discarded,
# because them cannot be bound to any key,
# ......

```

Which can be decoded, modified and written out with comment-preserving :

```go
package main

import (
	"log"
	"os"

	"github.com/GuanceCloud/toml"
)

type ReleaseHistory []struct {
	Date     string   `toml:"date"`
	Version  string   `toml:"version"`
	Features []string `toml:"features"`
}

type Author struct {
	Name          string     `toml:"name"`
	Age           int        `toml:"age"`
	Gender        string     `toml:"gender"`
	MaritalStatus bool       `toml:"marital status"`
	Hobbies       [][]string `toml:"hobbies"`
}

type Country struct {
	Name       string   `toml:"name"`
	Code       string   `toml:"code"`
	Continent  string   `toml:"continent"`
	MainCities []string `toml:"main_cities"`
	State      []State  `toml:"state"`
}

type State struct {
	Name       string    `toml:"name"`
	Capital    string    `toml:"capital"`
	Area       int       `toml:"area"`
	Population int       `toml:"population"`
	Latitude   [2]string `toml:"latitude"`
	Longitude  [2]string `toml:"longitude"`
}

type tomlStruct struct {
	Version        string         `toml:"version"`
	ReleaseHistory ReleaseHistory `toml:"release_history"`
	Author         Author         `toml:"author"`
	Country        []Country      `toml:"country"`
}

func main() {
	var ts tomlStruct

	meta, err := toml.DecodeFile("input.toml", &ts)
	if err != nil {
		log.Fatal(err)
	}

	ts.Author.Hobbies = append(ts.Author.Hobbies, []string{"volleyball"})
	ts.Version = "v1.2.3"
	ts.Author.Name = "GuanceCloud"
	ts.Country[1].State[0].Capital = "Boston"

	ts.Country[0].State = append(ts.Country[0].State, State{
		Name:       "Anhui",
		Capital:    "Hefei",
		Area:       140100,
		Population: 61270000,
		Latitude:   [2]string{"29°41′ N", "34°38′ N"},
		Longitude:  [2]string{"114°54′ E", "119°37′ E"},
	})

	enc := toml.NewEncoder(os.Stdout)
	if err := enc.EncodeWithComments(ts, meta); err != nil {
		log.Fatal(err)
	}
}

```

The output will be like this: 
```toml
# preserving comment example
# bind rules:
# 1. document comments will be bound to below nearest key
# 2. comment at the end of line will be bound to prefix key
version = "v1.2.3" # bound to version key


[[release_history]] # comment for release_history[0]
  date = "2023-03-01"
  version = "v0.3"
  features = ["1. [SESSION] Supports Sysmon for easy monitoring of system resources usage, such as CPU, memory, network", "2. Start WindTerm and select the profiles directory and quit.",  # comment for features[2]
"3. [TAB] Restore the last modified tab name. #626" # comments for features[3]
]

# this comment will be bound to below reeases[1]
[[release_history]] # this comment will also be bound to releases[1]
  date = "2023-02-01"
  version = "v0.2"

[[release_history]] # comment for release[2]
  date = "2023-01-01"
  version = "v0.1"

[author]
  name = "GuanceCloud" # name

  age = 32 # age

  gender = "male" # gender

  "marital status" = true # marital status


  # document comment for hobbies
  hobbies = [["basketball",  # 篮球
"football" # 足球
], ["tennis"],  # 网球

    # this is full line comment for sub array
["Snooker",  # sub array[0]
"Billy Billiards",  # sub array[1]
"Pyramid" # sub array[2]
],  # this is line tail comment for sub array
["volleyball"]] #comment for hobbies # also comment for hobbies


# the comment for first country
[[country]]
  name = "China" # comment bound to China

  code = "CN" # comment bound to CN

  continent = "Asia" # comment bound to Asia


  # main city list
  main_cities = ["Beijing",  # bounds to main_cities[0]
"Shanghai",  # bounds to main_cities[1]
"Guangzhou" # bounds to main_cities[2]
] # main city list of China over


  [[country.state]]
    name = "Anhui"
    capital = "Hefei"
    area = 140100
    population = 61270000
    latitude = ["29°41′ N", "34°38′ N"]
    longitude = ["114°54′ E", "119°37′ E"]

# the second country: America
# The United States of America (U.S.A. or USA), commonly known as the United States (U.S. or US) or America,
# is a country primarily located in North America. It consists of 50 states, a federal district,
# five major unincorporated territories, nine Minor Outlying Islands,[h] and 326 Indian reservations.
[[country]]

  # comment for America
  name = "America"
  code = "USA"
  continent = "North America"

  # the largest cities in America
  # including New York, Los Angeles ...
  #
  main_cities = ["New York",  # bounds to main_cities[0]: New York
"Los Angeles",  # bounds to main_cities[1]: Los Angeles
"Chicago",  # Chicago is an international hub for finance, culture, commerce, industry, education, technology, telecommunications, and transportation
"San Francisco"] # main city list of America over


  # this comment will be bound to country[1].state[0]
  [[country.state]]
    name = "Massachusetts"
    capital = "Boston"
    area = 21000
    population = 7029917
    latitude = ["41°14′ N",  # latitude from
"42°53′ N" # latitude to
]

    # comment for Longitude
    longitude = ["69°56′ W",  # Longitude from
"73°30′ W" # Longitude to
]

  [[country.state]] # this is comment for country[1].state[1]
    name = "Texas"
    capital = "Austin"
    area = 696241
    population = 29145505
    latitude = ["25°50′ N", "36°30′ N"]
    longitude = ["93°31′ W", "106°39′ W"]

[[country]]
  name = "Germany"
  code = "DE"
  continent = "Europe"
  main_cities = ["Hamburg", "Munich", "Frankfurt"] # comment bound to main_cities


```


### More complex usage
See the [`_example/`](/_example) directory for more example.
