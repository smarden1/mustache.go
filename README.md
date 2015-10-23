{{ mustache! }}

## Overview

This is an implementation of mustache written in go. [Mustache](http://mustache.github.com/mustache.5.html) is a logic-less templating language with a variety of implementations.

## Motivation

The main reason for writing yet another implementation was to learn go. This was my first project in go and as such, may not be completely idiomatic. It does however, pass almost all of the mustache specs, but it has not been used in a large project.

## Usage

  ```
  // Render will render a template using the provided data.
  Render(template string, data ...interface{}) (string, err)
  ```


  ```
  // Compile will compile a template. Compiled templates are faster if you use them more then once,
  // otherwise prefer Render.
  Compile(template string) (*Template, error)
  ```


  ```
  // Render will render a template using the provided data.
  (t *Template) Render(c ...interface{}) string
  ```

## TODOs

1. add lambda support
2. fix the remaining 3 partial specs that do not match