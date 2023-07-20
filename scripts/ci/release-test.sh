#!/usr/bin/env bash

# release.sh is the release script to make Okteto CLI versions publicly
# available.
#
# This scripts is meant to be executed by circleci and makes a few assumptions about
# the environment it runs on. It assumes the golangci executor which has a few
# binaries required by this script.
# TODO: parameterize this script to make it able to run locally
#
# Releases can be pulled from three distinct channels: stable, beta and dev.
#
# Releases are represented by git annotated tags in semver version and have their
# respective Github release.
# Releases from the stable channel have the format: MAJOR.MINOR.PATCH
# Releases from the beta channel have the MAJOR.MINOR.PATCH-beta.n
# Releases from the dev channel have the MAJOR.MINOR.PATCH-dev.n
#

# run in a subshell
{ (

        set -e # make any error fail the script
        set -u # make unbound variables fail the script

        # SC2039: In POSIX sh, set option pipefail is undefined
        # shellcheck disable=SC2039
        set -o pipefail # make any pipe error fail the script

        # RELEASE_TAG is the release tag that we want to release
        RELEASE_TAG="2.18.0"

        # RELEASE_COMMIT is the commit being released
        RELEASE_COMMIT="485b2be6406313888e0b6776ef098fba5a750864"

        # PSEUDO_TAG is the short sha that will be used to release in the case
        # that a release tag is not provided
        PSEUDO_TAG="485b2be"

        # CURRENT_BRANCH is the branch being released.
        CURRENT_BRANCH="HEAD"

        # REPO_OWNER is the owner of the repo (okteto)
        REPO_OWNER="okteto"

        # REPO_NAME is the name of the repo (okteto)
        REPO_NAME="okteto"

        # BIN_PATH is where the artifacts are stored. Usually precreated by the circleci
        # workflow.
        BIN_PATH="$PWD/artifacts/bin"

        ################################################################################
        # Sanity check
        ################################################################################

        if ! semver --version >/dev/null 2>&1; then
                echo "Binary \"semver\" does is required from running this scripts"
                echo "Please install it and try again:"
                echo "  $ curl -o /usr/local/bin/semver https://raw.githubusercontent.com/fsaintjacques/semver-tool/master/src/semver"
                exit 1
        fi

        if ! command -v okteto-ci-utils >/dev/null 2>&1; then
                echo "Binary \"okteto-ci-utils\" from the golangci circleci executor is required to run this script"
                echo "If you are running this locally you can go into the golang-ci executor repo and build the script from source:"
                echo "  $ go build -o /usr/local/bin/okteto-ci-utils ."
                exit 1
        fi

        if [ ! -d "$BIN_PATH" ]; then
                echo "Release artifacts should be stored in $BIN_PATH but the directory does no exist"
                exit 1
        elif [ -z "$(ls -A "$BIN_PATH")" ]; then
                echo "Release artifacts should be stored in $BIN_PATH but the directory is empty"
                exit 1
        fi

        if [ -z "$GITHUB_TOKEN" ]; then
                echo "GITHUB_TOKEN envvar not provided. It is required to create the Github Release and the release notes"
                exit 1
        fi

        echo "Releasing tag '${RELEASE_TAG}' from branch '${CURRENT_BRANCH}' at ${RELEASE_COMMIT}"

        ################################################################################
        # Resolve release channel
        ################################################################################

        # Resolve the channel from the tag that is being released
        # If the channel is unknown the release will fail
        CHANNELS=

        is_oficial_release=false

        # dev releases don't have tags
        if [ "$RELEASE_TAG" = "" ]; then
                CHANNELS=("dev")
        else
                beta_prerel_regex="^beta\.[0-9]+"
                prerel="$(semver get prerel "${RELEASE_TAG}")"

                # Stable releases are added to all channel
                if [ -z "$prerel" ]; then
                        CHANNELS=("stable" "beta" "dev")
                        is_oficial_release=true
                elif [[ $prerel =~ $beta_prerel_regex ]]; then
                        CHANNELS=("beta" "dev")

                else
                        echo "Unknown tag"
                        echo "Expected one of: "
                        echo "  - stable: MAJOR.MINOR.PATCH "
                        echo "  - beta: MAJOR.MINOR.PATCH-beta.n"
                        echo "$RELEASE_TAG matches none"
                        exit 1
                fi
        fi

        BIN_BUCKET_ROOT="downloads.okteto.com/cli"
        
        for chan in "${CHANNELS[@]}"; do
                echo "---------------------------------------------------------------------------------"
                tag="${RELEASE_TAG:-"$PSEUDO_TAG"}"
                echo "Releasing ${tag} into ${chan} channel"

                ##############################################################################
                # Update downloads.okteto.com
                ##############################################################################

                # BIN_BUCKET_NAME is the name of the bucket where the binaries are stored.
                # Starting at Okteto CLI 2.0, all these binaries are publicly accessible at:
                # https://downloads.okteto.com/cli/<channel>/<tag>
                
                BIN_BUCKET_ROOT_WITH_CHAN="${BIN_BUCKET_ROOT}/${chan}"
                BIN_BUCKET_NAME="${BIN_BUCKET_ROOT_WITH_CHAN}/${tag}"

                # VERSIONS_BUCKET_FILENAME are all the available versions for a release channel.
                # This is also publicly accessible at:
                # https://downloads.okteto.com/cli/<channel>/versions
                VERSIONS_BUCKET_FILENAME="downloads.okteto.com/cli/${chan}/versions"

                # upload artifacts
                echo "Syncing artifacts from $BIN_PATH with $BIN_BUCKET_NAME"
                #gsutil -m rsync -r "$BIN_PATH" "gs://$BIN_BUCKET_NAME"

                # Get the current versions file and add the current version being released.
                # These are the versions publicly accessible from this channel.
                # It is important to have them sorted so that the last version from the list
                # is always the latest and we can keep pushing older tags for maintenance and
                # whatnot.
                version_file="$HOME/versions-${chan}"
                version_file_tmp=$(mktemp)
                #gsutil cat "gs://${VERSIONS_BUCKET_FILENAME}" >"$version_file_tmp"
                echo "Current version list for ${chan} channel (showing latest 10):"
                tail -n 10 "$version_file_tmp"

                printf "%s\n" "${tag}" >>"$version_file_tmp"

                # dont sort the dev channel. Not all tags are semver and it's
                # safe to assume linear history
                if [ "${chan}" = "dev" ]; then
                        awk '!seen[$0]++' "$version_file_tmp" >"${version_file}"
                else
                        awk '!seen[$0]++' "$version_file_tmp" | perl -pe 's/\-(?=beta)/~/' | sort -V | perl -pe 's/~/-/' >"${version_file}"
                fi

                echo "Added ${tag} to the version list"
                echo "New version list for ${chan} channel (showing latest 10):"
                tail -n 10 "${version_file}"

                # After sorting, if the latest tag is the current tag update the root path
                # with the current binaries
                latest="$(tail -n1 "${version_file}")"

                echo "latest -> $latest"

                if [ "$tag" = "$latest" ]; then
                        echo "tag equal to latest"
                        #gsutil -m rsync "gs://$BIN_BUCKET_NAME" "gs://$BIN_BUCKET_ROOT_WITH_CHAN"
                        
                        echo $is_oficial_release

                        if [ "$is_oficial_release" = true ] ; then
                                # upload artifacts to bucket root (gs://downloads.okteto.com/cli)
                                echo "Syncing artifacts from $BIN_PATH with $BIN_BUCKET_ROOT"
                                echo "All sync"
                                #gsutil -m rsync -r "$BIN_PATH" "gs://$BIN_BUCKET_ROOT"
                        fi
                fi

                #gsutil -m -h "Cache-Control: no-store" -h "Content-Type: text/plain" cp "${version_file}" "gs://${VERSIONS_BUCKET_FILENAME}"
                echo "${chan} channel updated with ${tag}"
        done

        

        echo "All done"
); }
