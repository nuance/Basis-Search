include $(GOROOT)/src/Make.inc

CMDS=\
	server/index \
	server/merger

PKGS=\
	util/varint \
	util/posting_list

all: make

make: $(addsuffix .install, $(PKGS)) $(addsuffix .make, $(CMDS))
clean: $(addsuffix .nuke, $(PKGS)) $(addsuffix .clean, $(CMDS))

%.install:
	$(MAKE) -C $* install

# compile all packages before any command
%.make: $(addsuffix .install, $(PKGS))
	$(MAKE) -C $*

# establish dependancies between packages
util/posting_list.install: util/varint.install

%.clean:
	$(MAKE) -C $* clean

%.nuke:
	$(MAKE) -C $* nuke

