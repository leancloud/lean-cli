package runtimes

import (
	"io/ioutil"
)

var routerPHPContent = `<?php
if (file_exists($_SERVER["DOCUMENT_ROOT"] . $_SERVER["REQUEST_URI"])) {
    return false;
} else {
    require "public/index.php";
}
`

func getPHPEntryScriptPath() (string, error) {
	f, err := ioutil.TempFile("", "leangine_router.php")
	defer f.Close()
	if err != nil {
		return "", err
	}
	_, err = f.WriteString(routerPHPContent)
	return f.Name(), err
}
