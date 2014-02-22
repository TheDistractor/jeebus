# JeeBus development mode: launches "gin" and compiles CS/Jade/Stylus as needed
# -jcw, 2014-02-19

fs = require 'fs'
http = require 'http'
path = require 'path'
{execFile,spawn} = require 'child_process'

# look for modules relative to the current directory, not relative to this file
moduleDir = (s) -> path.resolve 'node_modules', s

coffee = require moduleDir 'coffee-script'
convert = require moduleDir 'convert-source-map'
jade = require moduleDir 'jade'
stylus = require moduleDir 'stylus'

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
  p
  
installGin = ->
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
  
compileCoffeeScriptWithMap = (sourceCode, filename) ->
  compiled = coffee.compile sourceCode,
    filename: filename
    sourceMap: true
    inline: true
    literate: path.extname(filename) isnt '.coffee'
  comment = convert
    .fromJSON(compiled.v3SourceMap)
    .setProperty('sources', [filename]) 
    .toComment()
  "#{compiled.js}\n#{comment}\n"
  
compileIfNeeded = (srcFile) ->
  if /\.(coffee|coffee\.md|litcoffee|jade|styl)$/i.test srcFile
    srcExt = path.extname srcFile
    destExt = switch srcExt
      when '.jade' then '.html'
      when '.styl' then '.css'
      else              '.js'
    destFile = srcFile.slice(0, - srcExt.length) + destExt

    t = Date.now()
    saveResult = (data) ->
      n = data.length
      ms = Date.now() - t
      console.log "[node] compile #{srcFile} -> #{destExt} #{n}b #{ms} ms"
      fs.writeFileSync destFile, data

    try
      srcStat = fs.statSync srcFile
      destStat = fs.statSync destFile  if fs.existsSync destFile
      unless destStat?.mtime >= srcStat.mtime
        src = fs.readFileSync srcFile, encoding: 'utf8'
        switch srcExt
          when '.jade'
            saveResult do jade.compile src, filename: srcFile, pretty: true
          when '.styl'
            stylus.render src, { filename: srcFile }, (err, css) ->
              if err
                console.log '[node] stylus error', srcFile, err
              else
                saveResult css
          else
            saveResult compileCoffeeScriptWithMap src, path.basename srcFile
    catch err
      console.log '[node] cannot compile', srcFile, err
  else if /\.(html|js)$/i.test srcFile
    triggerRefresh true
  else if /\.(css)$/i.test srcFile
    triggerRefresh false

traverseDirs = (dir, cb) -> # recursive directory traversal
  stats = fs.statSync dir
  if stats.isDirectory()
    cb dir
    for f in fs.readdirSync dir
      traverseDirs path.join(dir, f), cb

watchDir = (root, cb) -> # recursive directory watcher
  traverseDirs root, (dir) ->
    fs.watch dir, (event, filename) ->
      file = path.join dir, filename
      cb event, file

createWatcher = (root) ->
  console.log " ", root
  traverseDirs root, (dir) ->
    # console.log 'd:', dir
    for f in fs.readdirSync dir
      compileIfNeeded path.join dir, f
    fs.watch dir, (event, filename) ->
      file = path.join dir, filename
      if fs.existsSync file
        compileIfNeeded file
      else
        # TODO: delete compiled file
  root

ready = -> # called once gin reports that it is actually listening on a port
  console.log '[node] watching for file changes in:'
  createWatcher settings.AppDir or './app'
  createWatcher settings.BaseDir or './base'
  createWatcher settings.CommonDir or './common'

parseSettings = (fn) ->
  map = {}
  if fs.existsSync fn
    for line in fs.readFileSync(fn, 'utf8').split '\n'
      line = line.trim()
      i = line.indexOf('=')
      if line[0] != '#' and i > 0
        x = []
        for s in line.slice(0, i).trim().split '_'
          x.push s.slice(0, 1).toUpperCase() + s.slice(1).toLowerCase()
        k = x.join ""
        v = line.slice(i+1).trim()
        map[k] = JSON.parse v
  map

# debounce the refresh calls, in case lots of them are generated at once
reload = undefined
timer = undefined

# this is defined after the initial scans from createWatcher have completed
triggerRefresh = (refresh) ->
  reload or= refresh
  clearTimeout timer
  timer = setTimeout ->
    port = settings.Port or 3000
    # send a "GET /reload?..." to broadcast it on all connected websockets
    http.get("http://localhost:#{port}/reload?#{reload}").on 'error', (e) ->
      console.log "Got error: ", e.stack
    reload = undefined
  , 100

# Start of devmode application code

settings = parseSettings 'settings.txt'

# assume "go" and "gin" have been installed properly
gin = runGin()

# else, try to install "gin" first
gin.on 'error', (err) ->
  fatal 'cannot launch "gin"', err  unless err.code is 'ENOENT'
  installGin()
  
gin.on 'exit', ->
  console.error '[node] gin exited'
  process .exit 1

process.on 'uncaughtException', (err) ->
  console.error '[node] exception:', err.stack
  gin.kill()
