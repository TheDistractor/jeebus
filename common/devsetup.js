(function() {
  var JEEBUS_ROOT, execFile, fatal, fs, goDir, installJeeBus, jbDir, path, readline, rl, _ref,
    __slice = [].slice;

  console.log('\nThis script sets up a fresh application area based on JeeBus.\nYou need to supply the name of a new directory to initialise.\nIt will be prepared with a minimal set of files and settings.\n');

  fs = require('fs');

  execFile = require('child_process').execFile;

  readline = require('readline');

  path = require('path');

  JEEBUS_ROOT = (_ref = process.env.JEEBUS_ROOT) != null ? _ref : 'github.com/jcw/jeebus';

  fatal = function() {
    var args, s;
    s = arguments[0], args = 2 <= arguments.length ? __slice.call(arguments, 1) : [];
    console.error('\n[node] fatal error:', s);
    if (args.length) {
      console.error.apply(console, args);
    }
    return process.exit(1);
  };

  if (!process.env.GOPATH) {
    fatal('GOPATH undefined, please make sure "go" has been installed');
  }

  goDir = process.env.GOPATH.split(':')[0];

  jbDir = "" + goDir + "/src/" + JEEBUS_ROOT;

  installJeeBus = function(done) {
    if (fs.existsSync(jbDir)) {
      return done();
    } else {
      console.log("Fetching jeebus from https://" + JEEBUS_ROOT);
      return execFile('go', ['get', JEEBUS_ROOT], function(err, sout, serr) {
        if ((err != null ? err.code : void 0) === 'ENOENT') {
          fatal('"go" not found - please install it first, see http://golang.org/');
        }
        if (err) {
          fatal('fetching failed', serr);
        }
        if (!fs.existsSync(jbDir)) {
          fatal('still cannot find jeebus');
        }
        return done();
      });
    }
  };

  rl = readline.createInterface({
    input: process.stdin,
    output: process.stdout
  });

  rl.question('Directory name? ', function(appDir) {
    var name, title;
    rl.close();
    name = path.basename(appDir);
    title = name.charAt(0).toUpperCase() + name.slice(1).toLowerCase();
    if (fs.existsSync(appDir)) {
      fatal('please enter the name of a nonexistent directory to initialise');
    }
    return installJeeBus(function() {
      var s;
      fs.mkdirSync(appDir);
      fs.mkdirSync("" + appDir + "/app");
      fs.writeFileSync("" + appDir + "/index.js", "require('" + jbDir + "/common/devmode');\n");
      fs.writeFileSync("" + appDir + "/settings.txt", "BASE_DIR = \"" + jbDir + "/base\"\nCOMMON_DIR = \"" + jbDir + "/common\"\n");
      fs.writeFileSync("" + appDir + "/package.json", "{\n  \"name\": \"" + name + "\",\n  \"description\": \"This is the new " + title + " application.\",\n  \"version\": \"0.0.1\",\n  \"main\": \"index.js\",\n  \"dependencies\": {\n    \"coffee-script\": \"*\",\n    \"convert-source-map\": \"*\",\n    \"jade\": \"*\",\n    \"stylus\": \"*\"\n  }\n}\n");
      fs.writeFileSync("" + appDir + "/main.go", "package main\n\nimport (\n    \"log\"\n    \"" + JEEBUS_ROOT + "\"\n)\n\nconst Version = \"0.0.1\"\n\nfunc init() {\n    log.SetFlags(log.Ltime) // only display HH:MM:SS time in log entries\n}\n\nfunc main() {\n    println(\"\\n" + title + "\", Version, \"/ JeeBus\", jeebus.Version)\n    jeebus.Run()\n}\n");
      s = fs.readFileSync("" + jbDir + "/app/index.html", "utf8");
      s = s.replace('<script src="/demo/demo.js"></script>', '<!-- your code -->');
      fs.writeFileSync("" + appDir + "/app/index.html", s);
      s = fs.readFileSync("" + jbDir + "/app/startup.js");
      fs.writeFileSync("" + appDir + "/app/startup.js", s);
      return console.log("\n" + title + " has been created. To start it up, enter this command:\n    cd " + appDir + " && node .\nThen open the web page at http://localhost:3000/ - that's it!\n");
    });
  });

}).call(this);

//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiIiwic291cmNlUm9vdCI6IiIsInNvdXJjZXMiOlsiZGV2c2V0dXAuY29mZmVlIl0sIm5hbWVzIjpbXSwibWFwcGluZ3MiOiJBQU9BO0FBQUEsTUFBQSx1RkFBQTtJQUFBLGtCQUFBOztBQUFBLEVBQUEsT0FBTyxDQUFDLEdBQVIsQ0FBWSxpTUFBWixDQUFBLENBQUE7O0FBQUEsRUFRQSxFQUFBLEdBQUssT0FBQSxDQUFRLElBQVIsQ0FSTCxDQUFBOztBQUFBLEVBU0MsV0FBWSxPQUFBLENBQVEsZUFBUixFQUFaLFFBVEQsQ0FBQTs7QUFBQSxFQVVBLFFBQUEsR0FBVyxPQUFBLENBQVEsVUFBUixDQVZYLENBQUE7O0FBQUEsRUFXQSxJQUFBLEdBQU8sT0FBQSxDQUFRLE1BQVIsQ0FYUCxDQUFBOztBQUFBLEVBY0EsV0FBQSxxREFBd0MsdUJBZHhDLENBQUE7O0FBQUEsRUFnQkEsS0FBQSxHQUFRLFNBQUEsR0FBQTtBQUNOLFFBQUEsT0FBQTtBQUFBLElBRE8sa0JBQUcsOERBQ1YsQ0FBQTtBQUFBLElBQUEsT0FBTyxDQUFDLEtBQVIsQ0FBYyx1QkFBZCxFQUF1QyxDQUF2QyxDQUFBLENBQUE7QUFDQSxJQUFBLElBQTBCLElBQUksQ0FBQyxNQUEvQjtBQUFBLE1BQUEsT0FBTyxDQUFDLEtBQVIsZ0JBQWMsSUFBZCxDQUFBLENBQUE7S0FEQTtXQUVBLE9BQU8sQ0FBQyxJQUFSLENBQWEsQ0FBYixFQUhNO0VBQUEsQ0FoQlIsQ0FBQTs7QUFxQkEsRUFBQSxJQUFBLENBQUEsT0FBYyxDQUFDLEdBQUcsQ0FBQyxNQUFuQjtBQUNFLElBQUEsS0FBQSxDQUFNLDREQUFOLENBQUEsQ0FERjtHQXJCQTs7QUFBQSxFQXdCQSxLQUFBLEdBQVEsT0FBTyxDQUFDLEdBQUcsQ0FBQyxNQUFNLENBQUMsS0FBbkIsQ0FBeUIsR0FBekIsQ0FBOEIsQ0FBQSxDQUFBLENBeEJ0QyxDQUFBOztBQUFBLEVBeUJBLEtBQUEsR0FBUSxFQUFBLEdBQUUsS0FBRixHQUFTLE9BQVQsR0FBZSxXQXpC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QixFQUEyQyxNQUEzQyxDQTlDSixDQUFBO0FBQUEsTUFnREEsQ0FBQSxHQUFJLENBQUMsQ0FBQyxPQUFGLENBQVUsdUNBQVYsRUFBbUQsb0JBQW5ELENBaERKLENBQUE7QUFBQSxNQWlEQSxFQUFFLENBQUMsYUFBSCxDQUFpQixFQUFBLEdBQUUsTUFBRixHQUFVLGlCQUEzQixFQUE2QyxDQUE3QyxDQWpEQSxDQUFBO0FBQUEsTUFtREEsQ0FBQSxHQUFJLEVBQUUsQ0FBQyxZQUFILENBQWdCLEVBQUEsR0FBRSxLQUFGLEdBQVMsaUJBQXpCLENBbkRKLENBQUE7QUFBQSxNQW9EQSxFQUFFLENBQUMsYUFBSCxDQUFpQixFQUFBLEdBQUUsTUFBRixHQUFVLGlCQUEzQixFQUE2QyxDQUE3QyxDQXBEQSxDQUFBO2FBc0RBLE9BQU8sQ0FBQyxHQUFSLENBQWUsSUFBQSxHQUVuQixLQUZtQixHQUVaLGlFQUZZLEdBR2xCLE1BSGtCLEdBR1YsNkVBSEwsRUF2RFk7SUFBQSxDQUFkLEVBUjhCO0VBQUEsQ0FBaEMsQ0EzQ0EsQ0FBQTtBQUFBIiwic291cmNlc0NvbnRlbnQiOlsiIyBKZWVCdXMgZGV2ZWxvcG1lbnQgbW9kZSBzZXR1cCwgY2hlY2sgYW5kIGluc3RhbGwgd2hhdGV2ZXIgaXMgbmVlZGVkXG4jIC1qY3csIDIwMTQtMDItMTlcblxuIyBUbyBiZSB1c2VkIGFzOlxuIyAgIGN1cmwgLU8gaHR0cHM6Ly9yYXcuZ2l0aHViLmNvbS9qY3cvamVlYnVzL21hc3Rlci9jb21tb24vZGV2c2V0dXAuanNcbiMgICBub2RlIGRldnNldHVwLmpzXG5cbmNvbnNvbGUubG9nICcnJ1xuXG4gIFRoaXMgc2NyaXB0IHNldHMgdXAgYSBmcmVzaCBhcHBsaWNhdGlvbiBhcmVhIGJhc2VkIG9uIEplZUJ1cy5cbiAgWW91IG5lZWQgdG8gc3VwcGx5IHRoZSBuYW1lIG9mIGEgbmV3IGRpcmVjdG9yeSB0byBpbml0aWFsaXNlLlxuICBJdCB3aWxsIGJlIHByZXBhcmVkIHdpdGggYSBtaW5pbWFsIHNldCBvZiBmaWxlcyBhbmQgc2V0dGluZ3MuXG5cbicnJ1xuXG5mcyA9IHJlcXVpcmUgJ2ZzJ1xue2V4ZWNGaWxlfSA9IHJlcXVpcmUgJ2NoaWxkX3Byb2Nlc3MnXG5yZWFkbGluZSA9IHJlcXVpcmUgJ3JlYWRsaW5lJ1xucGF0aCA9IHJlcXVpcmUgJ3BhdGgnXG5cbiMgdGhlIHJlcG9zaXRvcnkgZnJvbSB3aGljaCBqZWVidXMgaXMgZmV0Y2hlZCBjYW4gYmUgb3ZlcnJpZGVuIHdpdGggYW4gZW52IHZhclxuSkVFQlVTX1JPT1QgPSBwcm9jZXNzLmVudi5KRUVCVVNfUk9PVCA/ICdnaXRodWIuY29tL2pjdy9qZWVidXMnXG5cbmZhdGFsID0gKHMsIGFyZ3MuLi4pIC0+XG4gIGNvbnNvbGUuZXJyb3IgJ1xcbltub2RlXSBmYXRhbCBlcnJvcjonLCBzXG4gIGNvbnNvbGUuZXJyb3IgYXJncy4uLiAgaWYgYXJncy5sZW5ndGhcbiAgcHJvY2Vzcy5leGl0IDFcblxudW5sZXNzIHByb2Nlc3MuZW52LkdPUEFUSFxuICBmYXRhbCAnR09QQVRIIHVuZGVmaW5lZCwgcGxlYXNlIG1ha2Ugc3VyZSBcImdvXCIgaGFzIGJlZW4gaW5zdGFsbGVkJ1xuXG5nb0RpciA9IHByb2Nlc3MuZW52LkdPUEFUSC5zcGxpdCgnOicpWzBdXG5qYkRpciA9IFwiI3tnb0Rpcn0vc3JjLyN7SkVFQlVTX1JPT1R9XCJcblxuaW5zdGFsbEplZUJ1cyA9IChkb25lKSAtPlxuICBpZiBmcy5leGlzdHNTeW5jIGpiRGlyXG4gICAgZG9uZSgpXG4gIGVsc2VcbiAgICBjb25zb2xlLmxvZyBcIkZldGNoaW5nIGplZWJ1cyBmcm9tIGh0dHBzOi8vI3tKRUVCVVNfUk9PVH1cIlxuICAgIGV4ZWNGaWxlICdnbycsIFsnZ2V0JywgSkVFQlVTX1JPT1RdLCAoZXJyLCBzb3V0LCBzZXJyKSAtPlxuICAgICAgaWYgZXJyPy5jb2RlIGlzICdFTk9FTlQnXG4gICAgICAgIGZhdGFsICdcImdvXCIgbm90IGZvdW5kIC0gcGxlYXNlIGluc3RhbGwgaXQgZmlyc3QsIHNlZSBodHRwOi8vZ29sYW5nLm9yZy8nXG4gICAgICBmYXRhbCAnZmV0Y2hpbmcgZmFpbGVkJywgc2VyciAgaWYgZXJyXG4gICAgICBmYXRhbCAnc3RpbGwgY2Fubm90IGZpbmQgamVlYnVzJyAgdW5sZXNzIGZzLmV4aXN0c1N5bmMgamJEaXJcbiAgICAgIGRvbmUoKVxuXG5ybCA9IHJlYWRsaW5lLmNyZWF0ZUludGVyZmFjZVxuICAgICAgICBpbnB1dDogcHJvY2Vzcy5zdGRpblxuICAgICAgICBvdXRwdXQ6IHByb2Nlc3Muc3Rkb3V0XG5cbnJsLnF1ZXN0aW9uICdEaXJlY3RvcnkgbmFtZT8gJywgKGFwcERpcikgLT5cbiAgcmwuY2xvc2UoKVxuICBuYW1lID0gcGF0aC5iYXNlbmFtZSBhcHBEaXJcbiAgdGl0bGUgPSBuYW1lLmNoYXJBdCgwKS50b1VwcGVyQ2FzZSgpICsgbmFtZS5zbGljZSgxKS50b0xvd2VyQ2FzZSgpXG5cbiAgaWYgZnMuZXhpc3RzU3luYyBhcHBEaXJcbiAgICBmYXRhbCAncGxlYXNlIGVudGVyIHRoZSBuYW1lIG9mIGEgbm9uZXhpc3RlbnQgZGlyZWN0b3J5IHRvIGluaXRpYWxpc2UnXG5cbiAgaW5zdGFsbEplZUJ1cyAtPlxuICAgIGZzLm1rZGlyU3luYyBhcHBEaXJcbiAgICBmcy5ta2RpclN5bmMgXCIje2FwcERpcn0vYXBwXCJcblxuICAgIGZzLndyaXRlRmlsZVN5bmMgXCIje2FwcERpcn0vaW5kZXguanNcIixcbiAgICAgIFwiXCJcInJlcXVpcmUoJyN7amJEaXJ9L2NvbW1vbi9kZXZtb2RlJyk7XFxuXCJcIlwiXG5cbiAgICBmcy53cml0ZUZpbGVTeW5jIFwiI3thcHBEaXJ9L3NldHRpbmdzLnR4dFwiLCBcIlwiXCJcbiAgICAgIEJBU0VfRElSID0gXCIje2piRGlyfS9iYXNlXCJcbiAgICAgIENPTU1PTl9ESVIgPSBcIiN7amJEaXJ9L2NvbW1vblwiXFxuXG4gICAgXCJcIlwiXG5cbiAgICBmcy53cml0ZUZpbGVTeW5jIFwiI3thcHBEaXJ9L3BhY2thZ2UuanNvblwiLCBcIlwiXCJcbiAgICAgIHtcbiAgICAgICAgXCJuYW1lXCI6IFwiI3tuYW1lfVwiLFxuICAgICAgICBcImRlc2NyaXB0aW9uXCI6IFwiVGhpcyBpcyB0aGUgbmV3ICN7dGl0bGV9IGFwcGxpY2F0aW9uLlwiLFxuICAgICAgICBcInZlcnNpb25cIjogXCIwLjAuMVwiLFxuICAgICAgICBcIm1haW5cIjogXCJpbmRleC5qc1wiLFxuICAgICAgICBcImRlcGVuZGVuY2llc1wiOiB7XG4gICAgICAgICAgXCJjb2ZmZWUtc2NyaXB0XCI6IFwiKlwiLFxuICAgICAgICAgIFwiY29udmVydC1zb3VyY2UtbWFwXCI6IFwiKlwiLFxuICAgICAgICAgIFwiamFkZVwiOiBcIipcIixcbiAgICAgICAgICBcInN0eWx1c1wiOiBcIipcIlxuICAgICAgICB9XG4gICAgICB9XFxuXG4gICAgXCJcIlwiXG5cbiAgICBmcy53cml0ZUZpbGVTeW5jIFwiI3thcHBEaXJ9L21haW4uZ29cIiwgXCJcIlwiXG4gICAgICBwYWNrYWdlIG1haW5cblxuICAgICAgaW1wb3J0IChcbiAgICAgICAgICBcImxvZ1wiXG4gICAgICAgICAgXCIje0pFRUJVU19ST09UfVwiXG4gICAgICApXG5cbiAgICAgIGNvbnN0IFZlcnNpb24gPSBcIjAuMC4xXCJcbiAgICAgIFxuICAgICAgZnVuYyBpbml0KCkge1xuICAgICAgICAgIGxvZy5TZXRGbGFncyhsb2cuTHRpbWUpIC8vIG9ubHkgZGlzcGxheSBISDpNTTpTUyB0aW1lIGluIGxvZyBlbnRyaWVzXG4gICAgICB9XG5cbiAgICAgIGZ1bmMgbWFpbigpIHtcbiAgICAgICAgICBwcmludGxuKFwiXFxcXG4je3RpdGxlfVwiLCBWZXJzaW9uLCBcIi8gSmVlQnVzXCIsIGplZWJ1cy5WZXJzaW9uKVxuICAgICAgICAgIGplZWJ1cy5SdW4oKVxuICAgICAgfVxcblxuICAgIFwiXCJcIlxuXG4gICAgcyA9IGZzLnJlYWRGaWxlU3luYyBcIiN7amJEaXJ9L2FwcC9pbmRleC5odG1sXCIsIFwidXRmOFwiXG4gICAgIyBsZWF2ZSBvdXQgdGhlIGRlbW8gYXBwIGZyb20gSmVlQnVzXG4gICAgcyA9IHMucmVwbGFjZSAnPHNjcmlwdCBzcmM9XCIvZGVtby9kZW1vLmpzXCI+PC9zY3JpcHQ+JywgJzwhLS0geW91ciBjb2RlIC0tPidcbiAgICBmcy53cml0ZUZpbGVTeW5jIFwiI3thcHBEaXJ9L2FwcC9pbmRleC5odG1sXCIsIHNcbiAgXG4gICAgcyA9IGZzLnJlYWRGaWxlU3luYyBcIiN7amJEaXJ9L2FwcC9zdGFydHVwLmpzXCJcbiAgICBmcy53cml0ZUZpbGVTeW5jIFwiI3thcHBEaXJ9L2FwcC9zdGFydHVwLmpzXCIsIHNcbiAgXG4gICAgY29uc29sZS5sb2cgXCJcIlwiXG5cbiAgICAgICN7dGl0bGV9IGhhcyBiZWVuIGNyZWF0ZWQuIFRvIHN0YXJ0IGl0IHVwLCBlbnRlciB0aGlzIGNvbW1hbmQ6XG4gICAgICAgICAgY2QgI3thcHBEaXJ9ICYmIG5vZGUgLlxuICAgICAgVGhlbiBvcGVuIHRoZSB3ZWIgcGFnZSBhdCBodHRwOi8vbG9jYWxob3N0OjMwMDAvIC0gdGhhdCdzIGl0IVxuXG4gICAgXCJcIlwiXG4iXX0=
