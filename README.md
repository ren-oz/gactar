[![GitHub release (latest by date)](https://img.shields.io/github/v/release/asmaloney/gactar)](https://github.com/asmaloney/gactar/releases/latest) ![Build](https://github.com/asmaloney/gactar/actions/workflows/build.yaml/badge.svg) [![GitHub](https://img.shields.io/github/license/asmaloney/gactar)](LICENSE)

# gactar

`gactar` is a tool for creating and running [ACT-R](https://en.wikipedia.org/wiki/ACT-R) models using a declarative file format called _amod_.

## Proof-of-Concept

**This is a proof-of-concept.**

Currently, `gactar` will take an [_amod_ file](#amod-file-format) and generate code to run it on three different ACT-R implementations:

- [CCM PyACTR](https://github.com/asmaloney/CCM-PyACTR) (python) - a.k.a. _"ccm"_
- [pyactr](https://github.com/jakdot/pyactr) (python)
- [ACT-R](https://github.com/asmaloney/ACT-R) (lisp) - a.k.a. _"vanilla"_

`gactar` will work with the tutorial models included in the _examples_ directory. It doesn't handle a lot beyond what's in there - it only works with memory modules, not perceptual-motor ones, and does not yet work with environments - so _it's limited at the moment_.

Given that gactar in its early stages, the amod syntax may change dramatically based on use and feedback.

### What isn't implemented?

A lot! The big, obvious one is environments (and therefore the visual & motor modules). That's a big challenge and probably not worth tackling if there isn't sufficient interest in this initial proof of concept. Environments may even prove impossible given the way they are implemented in the three frameworks, but I haven't yet explored this too deeply.

If there is sufficient interest in this project, my strategy going forward would be to continue implementing examples included with the three implementations, adding capabilities as necessary and, when the implementations differ, raising issues for discussion. Once all the non-environment capabilities are implemented, then I would turn to the environment issue.

## Why?

1. Provides a human-readable, easy-to-understand, standard format to define basic ACT-R models.
1. Allows the easy exchange of models with other researchers
1. Opens the possibility of a library of models which will run on multiple implementation frameworks.
1. Abstracts away the "programming" to focus on writing and understanding models.
1. Restricts the model to a small language to prevent programming "outside the model" (no sneaking in extra calculations or control-flow!).
1. Runs the same model on multiple ACT-R implementation frameworks.
1. Provides a very simple setup for teaching environments - gactar is self-contained in one executable and uses a setup script to download the implementation frameworks.
1. Generates human-readable code with comments linking back to the amod file which is useful for learning the implementations and comparing them.
1. Parses chunks (including the `examples` in an amod file) to catch and report errors in a user-friendly manner.

   **Example #1 (invalid variable name)**

   ```
    match {
        goal [isMember: ?obj ? nil]
    }
    do {
        recall [property: ?ojb category ?]
    }
   ```

   The CCM Suite implementation _fails silently_ when given invalid variables which makes it difficult to catch errors & can result in incorrect output. Instead of ignoring the incorrect variable, gactar outputs a nice error message so it's obvious what the problem is:

   ```
   recall statement variable '?ojb' not found in matches for production 'initialRetrieval' (line 53)
   ```

   **Example #2 (invalid slot name)**

   ```
    match {
        goal [isMember: ?obj ? nil]
    }
    do {
        set goal.resutl to 'pending'
    }
   ```

   The CCM Suite implementation produces the following error:

   ```
   Traceback (most recent call last):
   File "/path/gactar_Semantic_Run.py", line 8, in <module>
    model.run()
   File "/path/CCMSuite3/ccm/model.py", line 254, in run
    self.sch.run()
   File "/path/CCMSuite3/ccm/scheduler.py", line 116, in run
    self.do_event(heapq.heappop(self.queue))
   File "/path/CCMSuite3/ccm/scheduler.py", line 161, in do_event
    result=event.func(*event.args,**event.keys)
   File "/path/CCMSuite3/ccm/lib/actr/core.py", line 64, in _process_productions
    choice.fire(self._context)
   File "/path/CCMSuite3/ccm/production.py", line 51, in fire
    exec(self.func, context, self.bound)
   File "<production-initialRetrieval>", line 2, in <module>
   File "/path/CCMSuite3/ccm/model.py", line 22, in __call__
    val = self.func(self.obj, *args, **keys)
   File "/path/CCMSuite3/ccm/lib/actr/buffer.py", line 60, in modify
    raise Exception('No slot "%s" to modify to "%s"' % (k, v))
   Exception: No slot "resutl" to modify to "pending"
   end...
   ```

   Instead, by adding validation, gactar produces a much better message:

   ```
   slot 'resutl' does not exist in chunk 'isMember' for match buffer 'goal' in production 'initialRetrieval' (line 52)
   ```

## Design Goals

1. amod syntax & semantics should be designed for humans to read & understand (i.e. should not require a programming background to grok).
1. Only provide one way to do something in the amod language - this helps when reading someone else's code and keep the parser simple.
1. gactar should be as simple as possible to set up, use, and understand.

## Contributing

For information on how to contribute (code, bug reports, ideas, or other resources), please see the [CONTRIBUTING](doc/CONTRIBUTING.md) doc.

## Setup

1. Although the `gactar` executable itself is compiled for each platform, it requires **python3** to run the setup and to run the _ccm_ and _pyactr_ implementations. **python3** needs to be somewhere in your `PATH` environment variable.

2. `gactar` requires one or more of the three implementations (_ccm_, _pyactr_, _vanilla_) be installed.

`gactar` uses a python virtual environment to keep all the required python packages, lisp files, and other implementation files in one place so it does not affect the rest of your system. For more information about the virtual environment see the [python docs](https://docs.python.org/3/library/venv.html).

### Setup Virtual Environment

1. Run `./scripts/setup.sh`
   This will do several things to set up your environment:

   - create a [virtual environment](https://docs.python.org/3/library/venv.html) for the project in a directory called `env`
   - download the [CCM Suite](https://github.com/asmaloney/CCM-PyACTR) & put its files in the right place
   - install [pyactr](https://github.com/jakdot/pyactr) using pip
   - download "vanilla" [ACT-R](https://github.com/asmaloney/ACT-R)
   - (macOS-only) download & install the [Steel Bank Common Lisp](https://www.sbcl.org/index.html) (sbcl) compiler
   - (macOS-only) compile the ACT-R lisp files

2. You will need to activate the virtual environment by running this in the terminal before you run `gactar`:

   ```sh
   source ./env/bin/activate
   ```

   If it activated properly, your command line prompt will start with `(env)`. If you want to deactivate it, run `deactivate`.

### Install SBCL Lisp Compiler

**Note:** On macOS, these steps are handled by running the [setup file](#setup-virtual-environment).

For now this is only automated on macOS because the required files are not easy to determine programmatically. It may be possible to improve this in the future for other operating systems.

1. We are using the [Steel Bank Common Lisp](https://www.sbcl.org/index.html) (sbcl) compiler. Download the correct version [from here](https://www.sbcl.org/platform-table.html) by finding your platform (OS and architecture) in the table and clicking its box. Put the file in the `env` directory and unpack it there.

2. To install it in our environment, change to the new directory it created (e.g. `sbcl-1.2.11-x86-64-darwin`) and run this command (setting the path to wherever the env directory is):

   ```sh
   INSTALL_ROOT=/path/to/gactar/env/ ./install.sh
   ```

3. Once it is successfully installed, go back to the 'env' directory and run the following command to compile the main actr files using the lisp compiler (setting the path to wherever the env directory is):
   ```sh
   export SBCL_HOME=/path/to/env/lib/sbcl; sbcl --script actr/load-single-threaded-act-r.lisp
   ```
   This will take a few moments to compile all the ACT-R files so it is ready to use.

## Build

If you want to build `gactar`, you will need [git](https://git-scm.com/) and the [go compiler](https://golang.org/) installed.

Then you just need to clone this repo:

```sh
git clone https://github.com/asmaloney/gactar
cd gactar
```

...and run the build command:

```
go build
```

This will create the `gactar` executable.

## Test

To run the built-in tests, from the top-level of the repo run:

```
go test ./...
```

## Usage

```
gactar [OPTIONS] [FILES...]
```

### Command Line Options

**--debug, -d**: turn on debugging output

**--ebnf**: output amod EBNF to stdout and quit

**--framework, -f** [string]: add framework - valid frameworks: all, ccm, pyactr, vanilla (default: [all])

**--interactive, -i**: run an interactive shell

**--port, -p** [number]: port to run the webserver on (default: 8181)

**--web, -w**: start a webserver to run in a browser

## Example Usage

These examples assume you have set up your virtual environment properly. See [setup](#setup) above.

### Write Generated Code To Files

```
(env)$ ./gactar examples/count.amod
gactar version v0.1.0
ccm: Using Python 3.9.7 from /Users/maloney/dev/CogSci/gactar/env/bin/python3
	- Generating code for examples/count.amod
	- written to ccm_count.py
pyactr: Using Python 3.9.7 from /Users/maloney/dev/CogSci/gactar/env/bin/python3
	- Generating code for examples/count.amod
	- written to pyactr_count.py
vanilla: Using SBCL 1.2.11 from /Users/maloney/dev/CogSci/gactar/env/bin/sbcl
	- Generating code for examples/count.amod
	- written to vanilla_count.lisp
```

This will generate code for all active frameworks in the directory you are running from.

You can choose which frameworks to use with `-f` like this:

```
./gactar -f ccm -f vanilla examples/count.amod
gactar version v0.1.0
ccm: Using Python 3.9.7 from /Users/maloney/dev/CogSci/gactar/env/bin/python3
	- Generating code for examples/count.amod
	- written to ccm_count.py
vanilla: Using SBCL 1.2.11 from /Users/maloney/dev/CogSci/gactar/env/bin/sbcl
	- Generating code for examples/count.amod
	- written to vanilla_count.lisp
```

### Run Interactively

```
(env)$ ./gactar -i
gactar version v0.1.0
Type 'help' for a list of commands.
To exit, type 'exit' or 'quit'.
pyactr: Using Python 3.9.7 from /Users/maloney/dev/CogSci/gactar/env/bin/python3
vanilla: Using SBCL 1.2.11 from /Users/maloney/dev/CogSci/gactar/env/bin/sbcl
ccm: Using Python 3.9.7 from /Users/maloney/dev/CogSci/gactar/env/bin/python3
> help
  exit:        exits the program
  frameworks:  choose frameworks to run (e.g. "ccm pyactr", "all")
  help:        exits the program
  history:     outputs your command history
  load:        loads a model: load [FILENAME]
  quit:        exits the program
  reset:       resets the current model
  run:         runs the current model: run [INITIAL STATE]
  version:     outputs version info
> load examples/count.amod
 model loaded
 examples:
       run [countFrom: 2 5 starting]
       run [countFrom: 1 7 starting]
> frameworks ccm
active frameworks: ccm
> run [countFrom: 2 5 starting]
   0.000 production_match_delay 0
   0.000 production_threshold None
   0.000 production_time 0.05
   0.000 production_time_sd None
   0.000 memory.error False
   0.000 memory.busy False
   0.000 memory.latency 0.05
   0.000 memory.threshold 0
   0.000 memory.maximum_time 10.0
   0.000 memory.record_all_chunks False
   0.000 retrieval.chunk None
   0.050 production None
   0.050 memory.busy True
   0.050 goal.chunk countFrom 2 5 counting
   0.100 retrieval.chunk count 2 3
   0.100 memory.busy False
   0.100 production increment
   0.150 production None
2
   0.150 memory.busy True
   0.150 goal.chunk countFrom 3 5 counting
   0.200 retrieval.chunk count 3 4
   0.200 memory.busy False
   0.200 production increment
   0.250 production None
3
   0.250 memory.busy True
   0.250 goal.chunk countFrom 4 5 counting
   0.300 retrieval.chunk count 4 5
   0.300 memory.busy False
   0.300 production increment
   0.350 production None
4
   0.350 memory.busy True
   0.350 goal.chunk countFrom 5 5 counting
   0.350 production stop
   0.400 retrieval.chunk count 5 6
   0.400 memory.busy False
   0.400 production None
5
   0.400 goal.chunk None
Total time:    3.350
 goal.chunk None
 memory.busy False
 memory.error False
 memory.latency 0.05
 memory.maximum_time 10.0
 memory.record_all_chunks False
 memory.threshold 0
 production None
 production_match_delay 0
 production_threshold None
 production_time 0.05
 production_time_sd None
 retrieval.chunk count 5 6
end...
> exit
```

You may choose which of the frameworks to run using the `frameworks` command.

Specifying frameworks on the command line will limit you to selecting those frameworks. For example this will make only `ccm` available in interactive mode:

```
./gactar -f ccm -i
```

### Run As Web Server

```
(env)$ ./gactar -w
ccm: Using Python 3.9.7 from /path/to/gactar/env/bin/python3
pyactr: Using Python 3.9.7 from /path/to/gactar/env/bin/python3
vanilla: Using SBCL 1.2.11 from /path/to/gactar/env/bin/sbcl
Serving gactar on http://localhost:8181
```

Open `http://localhost:8181` in your browser. You can run the default example simply by clicking **Run**. You can also:

- select another example using the **Load Example** button
- modify the amod code in the editor
- **Save** the amod code to a file
- **Load** the amod code from a file
- set a **Goal** to override the default goal in the examples
- once it's been run, browse the generated code using the tabs at the top of the code editor

![gactar Web Interface](doc/images/gactar-web.png)

The results (and any errors) will be shown on the right and the generated code that was used to run the model on each framework is shown in the editor tabs.

**Important Note:** This web server is only intended to be run locally. It should not be used to expose gactar to the internet. Because we are running code, a lot more checking and validation of inputs would be required before doing so.

## amod File Format

Here is an example of the file format:

```
==model==

// The name of the model (used when generating code and for error messages)
name: count

// Description of the model (currently output as a comment in the generated code)
description: 'This is a model which adds numbers. Based on the ccm u1_count.py tutorial.'

// Examples of starting goals to use when running the model
examples {
    [countFrom: 2 5 starting]
    [countFrom: 1 3 starting]
}

==config==

// Turn on logging by setting 'log_level' to 'min', 'info' (default), or 'detail'
gactar { log_level: 'detail' }

// Declare chunks and their layouts
chunks {
    [count: first second]
    [countFrom: start end status]
}

==init==

// Initialize the memory
memory {
    [count: 0 1]
    [count: 1 2]
    [count: 2 3]
    [count: 3 4]
    [count: 4 5]
}

// Default goal
goal [countFrom: 2 5 starting]

==productions==

// Name of the production
start {
    // Optional description
    description: 'Starting point - first production to match'

    // Buffers to match
    match {
        goal [countFrom: ?start ?end starting]
    }
    // Steps to execute
    do {
        recall [count: ?start ?]
        set goal to [countFrom: ?start ?end counting]
    }
}

increment {
    match {
        goal [countFrom: ?x !?x counting]
        retrieval [count: ?x ?next]
    }
    do {
        print ?x
        recall [count: ?next ?]
        set goal.start to ?next
    }
}

stop {
    match {
        goal [countFrom: ?x ?x counting]
    }
    do {
        print ?x
        clear goal
    }
}
```

You can find other examples of `amod` files in the [examples folder](examples).

### Special Chunks

User-defined chunks must not begin with '\_' or be named `goal`, `retrieval`, or `memory` - these are reserved for internal use. Currently there is one internal chunk - _\_status_ - which is used to check the status of buffers and memory.

It is used in a `match` as follows:

```
match {
    goal [_status: full]
    memory [_status: error]
}
```

For buffers, the valid statuses are `full` and `empty`.

For memory, valid statuses are `busy`, `free`, `error`.

### Pattern Syntax

The _match_ section matches _patterns_ to buffers. Patterns are delineated by square brackets - e.g. `[property: ?obj category ?cat]`. The first item is the chunk name and the items after the colon are the slots. These are parsed to ensure their format is consistent with _chunks_ which are declared in the _config_ section.

The _do_ section in the productions uses a small language which currently understands the following commands:

| command                                                                  | example                                 |
| ------------------------------------------------------------------------ | --------------------------------------- |
| **clear** _(buffer name)+_                                               | **clear** goal, retrieval               |
| **print** _(string or var or number)+_                                   | **print** 'text', ?var, 42              |
| **recall** _(pattern)_                                                   | **recall** [car: ?colour]               |
| **set** _(buffer name)_._(slot name)_ **to** _(string or var or number)_ | **set** goal.wall_colour **to** ?colour |
| **set** _(buffer name)_ **to** _(pattern)_                               | **set** goal **to** [start: 6 nil]      |

## Processing

This diagram shows how an amod file is processed by gactar. The partial paths at the bottom of the items is the path to the source code responsible for that part of the processing.

![How gactar processes an amod file](doc/images/gactar.svg)
