# JeeBus development mode: re-compiles CS/Jade/Stylus as needed
# -jcw, 2014-02-19

fs = require 'fs'
path = require 'path'
{spawn} = require 'child_process'

# look for modules relative to the current directory, not relative to this file
moduleDir = (s) -> path.resolve 'node_modules', s

coffee = require moduleDir 'coffee-script'
convert = require moduleDir 'convert-source-map'
jade = require moduleDir 'jade'
stylus = require moduleDir 'stylus'

fatal = (s) ->
  console.error '\n[node] fatal error:', s
  process.exit 1

main = undefined

runMain = ->
  args = ['run', ]
  for f in fs.readdirSync '.'
    args.push f  if path.extname(f) is '.go' and not /_test\./.test f
  console.log '[node] go', args.join ' '
  main = spawn 'go', args, stdio: ['ipc', process.stdout, process.stderr]
  main.on 'close', (code) ->
    fatal 'unexpected termination of "main", code: ' + code  if code > 0
  main.on 'error', (err) ->
    fatal 'cannot launch "go"'
  main.on 'exit', ->
    fatal 'main exited'
  
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
    main.send true # request a full page reload
  else if /\.(css)$/i.test srcFile
    main.send false # request a stylesheet reload

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

# Start of devmode application code --------------------------------------------

runMain()

process.on 'message', (msg) ->
  console.log 'got message:', msg
  
console.log '[node] watching for file changes in:'

try settings = require('./setup').settings
createWatcher settings?.AppDir or './app'
createWatcher settings?.BaseDir or './base'
createWatcher settings?.CommonDir or './common'
