# Notes

### Update the papermod theme

```
cd themes/PaperMod
git pull
```

### Local builds

Run the followwing to serve locally
```
 $ hugo serve --minify  --disableFastRender
```

I can then visit http://localhost:1313/ with a web browser to view the local changes.

### To publish

publishDir has been changed to docs to allow building locally & pushing the repo to github.  As 
new content will get generated in docs, this will get picked up by github pages & served.  To generate content 

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
 