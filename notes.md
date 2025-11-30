# Notes

### Update the papermod theme

```
git submodule update --remote --merge
```

### Local builds

Run the following to serve locally
```
 $ hugo serve
```

Visit http://localhost:1313/ with a web browser to view the local changes.

### To publish

publishDir has been changed to `docs` to allow building locally & pushing the repo to github.  As 
new content will get generated in `docs`, this will get picked up by github pages & served.  

To create a new page.  This is considered a [page bundle](https://gohugo.io/content-management/page-bundles/) 
where all elements are stored in the my_topic directory.

```
hugo new content/articles/my_topic/index.md
```

To publish content to github 

* run `hugo` without any args.  This will generate the content/
* git add, commit, push
* inspect changes at http://amf3.github.io

### Urls

* [Hugo Documenation](https://gohugo.io/documentation/)
* [PaperMod Wiki](https://github.com/adityatelange/hugo-PaperMod/wiki)
* [PaperMod Repo](https://github.com/adityatelange/hugo-PaperMod)
* [PaperMod Example Site](https://github.com/adityatelange/hugo-PaperMod/tree/exampleSite)
* [PaperMod Example Site Source](https://github.com/adityatelange/hugo-PaperMod/tree/exampleSite)
* [YAML TO TOML converter](https://transform.tools/yaml-to-toml)

### Manage Hugo with go modules

go mod init github.com/<myusername>/<myBlogName>
go get -tool github.com/gohugoio/hugo@v0.152.2
go tool hugo serve or whatever subcommand is needed
