```
                                  .    o8o                         
                                .o8    `"'                         
oo.ooooo.   .ooooo.   .oooo.o .o888oo oooo  ooo. .oo.    .oooooooo 
 888' `88b d88' `88b d88(  "8   888   `888  `888P"Y88b  888' `88b  
 888   888 888   888 `"Y88b.    888    888   888   888  888   888  
 888   888 888   888 o.  )88b   888 .  888   888   888  `88bod8P'  
 888bod8P' `Y8bod8P' 8""888P'   "888" o888o o888o o888o `8oooooo.  
 888                                                    d"     YD  
o888o                                                   "Y88888P'  
```
# Posting
Posting is a lightweight microblogging framework. I created it because I wanted an easy way to share personal photos without relying on social media. It is written in Go, uses Firestore as a database, and Imgur for image hosting. You can see an example of it in action [here](https://posting.website).

## Features
* Lightweight
* Supports text and images
* Very cheap (almost free) to operate at small scale
* Easy to publish new posts
* Easy to deploy (and well suited for Google App Engine)


## Running Locally
### Requirements
* Go
* Docker
* An Imgur account with corresponding [client ID](https://api.imgur.com/)

### Start local server
1. Edit `dev_vars.env` so that `IMGUR_CLIENT_ID` contains your Imgur client ID.
2. Start the local Firestore DB emulator via `make docker-firestore-run`
3. Start the server on port `8008` via `make run-local`
4. Navigate to http://localhost:8008/posts. You should see a mostly-empty homepage template.

### Add a new post
1. Head to http://localhost:8008/new where you will be greeted with a basic auth login prompt
    - The default username is `dev` and the password is `posting`.
    - You can modify these values in `dev_vars.env`; note that the password is base64 encoded.
2. Now you will find a minimal (ugly) HTML form for adding a new Post
    - Add a text blurb and optionally select an image to upload.
    - Note that a Post must contain a text blurb OR at least one image.
3. Hit the `POST` button. You will be redirected to the home page, where your post is visible.


## Deploy
Deploying a Posting instance is fairly simple. How and where you deploy is up to you, but these are the necessary steps.

1. Create a [Firestore Database](https://cloud.google.com/firestore/docs/create-database-server-client-library).
2. Configure your deployment environment with the appropriate environment variables.
3. Clone this repo to your deployment environment then `go build` and execute the resulting artifact.


I personally host Posting on Google App Engine because it is easy and cheap. Here is an example `app.yaml`:

```
runtime: go119
instance_class: F1

env_variables:
  IMGUR_CLIENT_ID: "your client ID"
  GCLOUD_PROJECT_ID: "your project ID"
  BASIC_AUTH_USERNAME: "changeme"
  BASIC_AUTH_PASSWORD: "some base64 encoded password"

handlers:
- url: /.*
  script: auto
  secure: always
```
