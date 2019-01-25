<div align="center">
    <img src="/paper/img/misc/thor-logo-with-text.png?raw=true?" alt="thor-logo" width="250">
    <h3><a style="color:#9900FF">T</a>ransforming <a style="color:#9900FF">H</a>ashed <a style="color:#9900FF">O</a>tus to <a style="color:#9900FF">R</a>gba</h3>
    <hr>
    <a href="https://travis-ci.org/will-rowe/thor"><img src="https://travis-ci.org/will-rowe/thor.svg?branch=master" alt="travis"></a>
</div>

***


>Under active development...

***

## Overview

`THOR` is a tool that will generate PNG images from genomic data, for use in machine learning applications.

So far, it will:

* [histosketch]() a bunch of FASTA files, creating a set of [HULK histosketches]()
* colour these histosketches to RGB values, so that each sketch of length *x* will be encoded into *x* RGB values
* build a PNG image from an OTU table where each row of pixels corresponds to a coloured histosketch 

It's a work in progress, but we've had some success in using these images in Neural Nets to classify the Human Microbiome Project 16S samples by body site.


## ToDo

When an OTU is not present in the refseq collection, it is ignored by THOR. This needs to be handled properly.




