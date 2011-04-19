include $(GOROOT)/src/Make.inc

PKGS=\
	util/varint \
	util/bufferpool \
	match/postinglist

all: make

make: $(addsuffix .install, $(PKGS)) $(addsuffix .make, $(CMDS))
clean: $(addsuffix .nuke, $(PKGS)) $(addsuffix .clean, $(CMDS))

%.install:
	$(MAKE) -C $* install

# compile all packages before any command
%.make: $(addsuffix .install, $(PKGS))
	$(MAKE) -C $*

# establish dependencies between packages
match/postinglist.install: util/varint.install

%.clean:
	$(MAKE) -C $* clean

%.nuke:
	$(MAKE) -C $* nuke

