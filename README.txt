Simple utility to interactively choose file/directory on the command line.

Chosen file/directory is output on the stdout, or non-zero exit status is returned.

git get -u github.com/indoles/gopick

bash example:

function mycd() {
  local d=$($GOPATH/bin/gopick)
	if [[ $? ]]; then
		 cd $d
  fi
}
