# JeeBus development mode: launch "gin" and compile CS/Jade/Stylus as needed
# -jcw, 2014-02-19

{execFile,spawn} = require 'child_process'

fatal = (s, args...) ->
  console.error '\n[node] fatal error:', s
  console.error args...  if args.length
  process.exit 1

runGin = (done) ->
  p = spawn 'gin', [], stdio: 'pipe'
  p.on 'close', (code) ->
    fatal 'unexpected termination of "gin", code: ' + code  if code > 0
  p.stdout.on 'data', (data) ->
    s = data.toString()
    process.stdout.write s if data.length > 0
    ready()  if /listening on port/.test s
  p.stderr.on 'data', (data) ->
    s = data.toString()
    process.stderr.write s  unless /execvp\(\)/.test s
  return p
  
ready = ->
  console.log '[node] watching for file changes (NOT YET)'
  
# assume "go" and "gin" have been installed properly
gin = runGin()

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
    gin = runGin()
    gin.on 'error', (err) ->
      fatal 'still cannot launch "gin" - is $GOPATH/bin in your $PATH?'
