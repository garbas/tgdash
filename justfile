# Record demo GIF
demo:
    vhs demo/demo.tape

# Build the site
site:
    rm -rf site/public
    cd site && hugo
