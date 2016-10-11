// Copyright © 2016 Alexander Gugel <alexander.gugel@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package registry implements a CommonJS compliant package registry.
// See http://wiki.commonjs.org/wiki/Packages/Registry
package registry

import (
    "crypto/sha1"
    "encoding/hex"
    log "github.com/Sirupsen/logrus"
    "github.com/alexanderGugel/nerva/storage"
    "github.com/alexanderGugel/nerva/util"
    "github.com/libgit2/git2go"
    "regexp"
)

// PackageRoot represents a CommonJS package root document containing all
// available versions of a given package available in its local repository.
// The root object that describes all versions of a package MUST be a JSON
// object with the following fields:
// name: The name of the package. When both are decoded, this MUST match the
// “package name” portion of the URL. That is, packages with irregular
// characters in their names would be URL-Encoded in the request, and
// JSON-encoded in the data. So, a request to
// /%C3%A7%C2%A5%C3%A5%C3%B1%C3%AE%E2%88%82%C3%A9 would show a package root
// object with “\u00e7\u00a5\u00e5\u00f1\u00ee\u2202\u00e9” as the name, and
// would refer to the “ç¥åñî∂é” project.
// versions: An object hash of version identifiers to valid “package version
// url” responses: either URL strings or package descriptor objects.
// See http://wiki.commonjs.org/wiki/Packages/Registry#Package_Root_Object
type PackageRoot struct {
    Name     string               `json:"name"`
    DistTags *PackageDistTags     `json:"dist-tags"`
    Versions *PackageRootVersions `json:"versions"`
}

// PackageDistTags represents the dist-tags of a package root object. It maps
// Common JS tags, such latest, to specific versions.
type PackageDistTags map[string]string

// PackageDist describes how a package can be downloaded.
type PackageDist struct {
    Tarball string `json:"tarball"`
    Shasum  string `json:"shasum"`
}

var versionTagRef = regexp.MustCompile("^refs\\/tags\\/v(.*)$")

// NewPackageRoot creates a new CommonJS package root document.
func NewPackageRoot(name string, url string, repo *git.Repository, shaCache *ShaCache) (*PackageRoot, error) {
    versions := PackageRootVersions{}
    contextLog := log.WithFields(log.Fields{"name": name})

    latest := ""

    repo.Tags.Foreach(func(tagRef string, id *git.Oid) error {
        contextLog := contextLog.WithFields(log.Fields{"tagRef": tagRef})

        if !versionTagRef.MatchString(tagRef) {
            contextLog.Debug("skipping non-version tag")
            return nil
        }

        packageVersion, err := NewPackageVersion(repo, id)
        contextLog = contextLog.WithFields(log.Fields{"packageVersion": packageVersion})
        if err != nil || packageVersion == nil {
            util.LogErr(contextLog, err, "failed to generate package version")
            return nil
        } else if version, ok := (*packageVersion)["version"].(string); ok {
            contextLog = contextLog.WithFields(log.Fields{"version": version})
            if versions[version] != nil {
                contextLog.Warn("duplicate version")
            }
            tarball := "http://" + url + "/" + name + "/-/" + id.String()

            shasum, ok := shaCache.Get(*id)
            if !ok {
                hasher := sha1.New()
                d, err := storage.NewDownload(repo, id)
                if err != nil || d == nil {
                    util.LogErr(contextLog, err, "failed to create download")
                    return nil
                }
                if err := d.Start(hasher); err != nil {
                    util.LogErr(contextLog, err, "failed to start download")
                }
                shasum = hex.EncodeToString(hasher.Sum(nil))
            }

            shaCache.Add(*id, shasum)
            (*packageVersion)["dist"] = &PackageDist{tarball, shasum}
            versions[version] = packageVersion
            latest = version
        }
        return nil
    })

    distTags := PackageDistTags{}
    if latest != "" {
        distTags["latest"] = latest
    }
    packageRoot := &PackageRoot{name, &distTags, &versions}
    return packageRoot, nil
}
