// Copyright 2021 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Run to generate func-e man page
package main 

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/tetratelabs/func-e/internal/globals"
	"github.com/tetratelabs/func-e/internal/cmd"
)

func main() {
	path := flag.String("p", ".", "Path to write man page to.")
	flag.Parse()

	app := cmd.NewApp(&globals.GlobalOpts{})

	manpage, err := app.ToMan()
	if err != nil {
		fmt.Printf("Unable to convert cli app to man page: %v\n", err)
	}

	clean_path := filepath.Clean(*path + "/func-e.8")

	err = ioutil.WriteFile(clean_path, []byte(manpage), 0777)
	if err != nil {
		fmt.Printf("Unable to write man page: %v\n", err)
	}
}
