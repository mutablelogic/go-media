#!/bin/sh
# brewfile.sh prints a Brewfile of the Homebrew formulae which the libraries
# installed in a prefix link with, from their pkg-config files, to install
# with brew bundle on another Mac:
#
#   etc/scripts/brewfile.sh build/install > Brewfile
#
# Each library they link with is looked for in the directories they name,
# then in Homebrew's, and is from the formula whose keg it's in. Libraries
# which aren't from Homebrew, such as the prefix's own and the system's,
# aren't listed.
set -e

if [ $# -ne 1 ]; then
	echo "usage: $0 prefix" >&2
	exit 2
fi
brew=$(brew --prefix)
pkgconfig=$(cd "$1" && pwd)/lib/pkgconfig
modules=$(cd "$pkgconfig" && ls *.pc | sed 's/\.pc$//')
flags=$(PKG_CONFIG_PATH="$pkgconfig" ${PKG_CONFIG:-pkg-config} --static --libs $modules)

# The directories to look in, and the libraries to look for
dirs=""
libs=""
for flag in $flags; do
	case "$flag" in
	-L*) dirs="$dirs ${flag#-L}" ;;
	-l*) libs="$libs ${flag#-l}" ;;
	esac
done
dirs="$dirs $brew/lib"

# The formula of each library, and pkg-config, to find them with
{
	echo pkgconf
	for lib in $libs; do
		for dir in $dirs; do
			for f in "$dir/lib$lib.dylib" "$dir/lib$lib.a"; do
				[ -e "$f" ] || continue
				path=$(realpath "$f")
				case "$path" in
				"$brew/Cellar/"*)
					formula=${path#"$brew/Cellar/"}
					echo "${formula%%/*}"
					;;
				esac
				break 2
			done
		done
	done
} | sort -u | sed 's/.*/brew "&"/'
