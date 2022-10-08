### Setup

Linnea supports running in two modes, either independent or with support from AWS Lambda

Running in independent mode is recommended as it is easier.

The reason for integrating AWS lambdas is for fetching images faster by directly allowing the lambda to download images from a S3 bucket.

Linnea will do the rest, if you want to use a Lambda read up to step 3 and then open the [AWS setup](./aws-setup.md) guide.

### Step 1 - Install Linnea

Install and Setup [Golang](https://go.dev/dl/)<br />

Then

```bash
git clone https://github.com/JoachimFlottorp/linnea.git
cd linnea
make build
```

### Step 2 - Setup dependecies

Copy the <b>config.default.json</b> to <b>config.json</b> and fill out the values

Install and Setup [Redis](https://redis.io/)<br />
Setup [AWS S3](https://aws.amazon.com/s3/)<br />

#### Step 2.1 - Setup Twitch

Twitch is used for authenticating users, as it's easier to leech of their authentication system.

Create a Twitch application [here](https://dev.twitch.tv/console/apps/create)
And set <i>OAuth Redirect URLs</i> to <i>https://YOUR_URL/auth/callback</i>

Set the category to <i>Website Integration</i> and copy over the <i>Client ID</i> and <i>Client Secret</i> to the <i>config.json</i> file.

### Step 2.2 - Generate JWT Secret

```bash
openssl rand -base64 64
```

Place in <i>config.json</i>.http.jwt.secret

### Step 3 - Run Linnea

```bash
make run
```
