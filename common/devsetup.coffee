# JeeBus development mode setup, check and install whatever is needed
# -jcw, 2014-02-19

# To be used as:
#   curl -O https://raw.github.com/jcw/jeebus/master/common/devsetup.js
#   node devsetup.js

console.log '''

  This script sets up a fresh application area based on JeeBus.
  You need to supply the name of a new directory to initialise.
  It will be prepared with a minimal set of files and settings.

'''

fs = require 'fs'
{execFile} = require 'child_process'
readline = require 'readline'
path = require 'path'

# the repository from which jeebus is fetched can be overriden with an env var
JEEBUS_ROOT = process.env.JEEBUS_ROOT ? 'github.com/jcw/jeebus'

fatal = (s, args...) ->
  console.error '\n[node] fatal error:', s
  console.error args...  if args.length
  process.exit 1

unless process.env.GOPATH
  fatal 'GOPATH undefined, please make sure "go" has been installed'

goDir = process.env.GOPATH.split(':')[0]
jbDir = "#{goDir}/src/#{JEEBUS_ROOT}"

installJeeBus = (done) ->
  if fs.existsSync jbDir
    done()
  else
    console.log "Fetching jeebus from https://#{JEEBUS_ROOT}"
    execFile 'go', ['get', JEEBUS_ROOT], (err, sout, serr) ->
      if err?.code is 'ENOENT'
        fatal '"go" not found - please install it first, see http://golang.org/'
      fatal 'fetching failed', serr  if err
      fatal 'still cannot find jeebus'  unless fs.existsSync jbDir
      done()

rl = readline.createInterface
        input: process.stdin
        output: process.stdout

rl.question 'Directory name? ', (appDir) ->
  rl.close()
  name = path.basename appDir
  title = name.charAt(0).toUpperCase() + name.slice(1).toLowerCase()

  if fs.existsSync appDir
    fatal 'please enter the name of a nonexistent directory to initialise'

  installJeeBus ->
    fs.mkdirSync appDir
    fs.mkdirSync "#{appDir}/app"

    fs.writeFileSync "#{appDir}/index.js",
      """require('#{jbDir}/common/devmode');\n"""

    fs.writeFileSync "#{appDir}/settings.txt", """
      BASE_DIR = "#{jbDir}/base"
      COMMON_DIR = "#{jbDir}/common"\n
    """

    fs.writeFileSync "#{appDir}/package.json", """
      {
        "name": "#{name}",
        "description": "This is the new #{title} application.",
        "version": "0.0.1",
        "main": "index.js"
      }\n
    """

    fs.writeFileSync "#{appDir}/main.go", """
      package main

      import (
          "log"
          "#{JEEBUS_ROOT}"
      )

      const Version = "0.0.1"
      
      func init() {
          log.SetFlags(log.Ltime) // only display HH:MM:SS time in log entries
      }

      func main() {
          println("\\n#{title}", Version, "/ JeeBus", jeebus.Version)
          jeebus.Run()
      }\n
    """

    s = fs.readFileSync "#{jbDir}/app/index.html", "utf8"
    # leave out the demo app from JeeBus
    s = s.replace '<script src="/demo/demo.js"></script>', '<!-- your code -->'
    fs.writeFileSync "#{appDir}/app/index.html", s
  
    s = fs.readFileSync "#{jbDir}/app/startup.js"
    fs.writeFileSync "#{appDir}/app/startup.js", s
  
    console.log """

      #{title} has been created. To start it up, enter this command:
          cd #{appDir} && node .
      Then open the web page at http://localhost:3000/ - that's it!

    """
