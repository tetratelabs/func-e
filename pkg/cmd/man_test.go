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

package cmd

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/tetratelabs/func-e/pkg/globals"
	"github.com/tetratelabs/func-e/pkg/moreos"
)

const siteManpageFile = "../../packaging/nfpm/func-e.8"

func TestManPageMatchesCommands(t *testing.T) {
	if runtime.GOOS == moreos.OSWindows {
		t.SkipNow()
	}

	app := NewApp(&globals.GlobalOpts{})

	expected, err := app.ToMan()
	require.NoError(t, err)

	actual, err := os.ReadFile(siteManpageFile)
	require.NoError(t, err)
	require.Equal(t, expected, string(actual))
}
