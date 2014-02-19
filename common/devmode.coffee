# JeeBus development mode: launch "gin" and compile CS/Jade/Stylus as needed
# -jcw, 2014-02-19

{execFile,spawn} = require 'child_process'

fatal = (s, args...) ->
  console.error '\n[node] fatal error:', s
  console.error args...  if args.length
  process.exit 1

# assume "go" and "gin" have been installed properly
gin = spawn 'gin', [], stdio: [process.stdin, process.stdout, 'pipe']

# else, try to install "gin" first
gin.on 'error', (err) ->
  fatal 'cannot launch "gin"', err  unless err.code is 'ENOENT'
  console.log '"gin" tool not found, installing...'
  
  # assume "go" has been installed properly
  execFile 'go', ['get', 'github.com/codegangsta/gin'], (err, sout, serr) ->
    
    # installing "go" cannot be done automatically
    if err
      if err.code is 'ENOENT'
        fatal '"go" not found - please install it first, see http://golang.org/'
      fatal 'install of "gin" failed', serr

    # ok, try running "gin" again
    gin = spawn 'gin', [], stdio: [process.stdin, process.stdout, 'pipe']
    gin.on 'error', (err) ->
      fatal 'still cannot launch "gin" - is $GOPATH/bin in your $PATH?'

# don't expect "gin" to ever terminate
gin.on 'close', (code) ->
  fatal 'unexpected termination of "gin", code: ' + code  if code > 0

watcher = ->
  console.log '[node] watching for file changes (NOT YET)'

setTimeout watcher, 100
