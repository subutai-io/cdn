package app

import "github.com/subutai-io/cdn/libgorjun"

var (
	Integration int
)

var (
	Apt       = "https://sysnetcdn.subutai.io:8338/kurjun/rest/apt/"
	Raw       = "https://sysnetcdn.subutai.io:8338/kurjun/rest/raw/download?id="
	Template  = "https://sysnetcdn.subutai.io:8338/kurjun/rest/template/download?id="
	FilesDir  = "/tmp/cdn-test-data/files/"
	Localhost = "http://127.0.0.1:8080"
)

var (
	Subutai = gorjun.VerifiedGorjunUser()
	Lorem   = gorjun.FirstGorjunUser()
	Ipsum   = gorjun.SecondGorjunUser()
)

var (
	PublicScope  = 0
	PrivateScope = 1
)

var (
	IDsLayer   = 0
	NamesLayer = 1
)

// all uploaded files
var (
	// 0-layer stands for IDs of public files
	// 1-layer stands for IDs of private files
	UserFiles = []map[string][]string{
		{
			"subutai": {

			},
			"lorem":   {

			},
			"ipsum":   {

			},
		},
		{
			"subutai": {

			},
			"lorem":   {

			},
			"ipsum":   {

			},
		},
	}
)

var (
	Repos = []map[string][]string{
		{
			"subutai": {
				"raw", "template",
			},
			"lorem":   {
				"raw", "template", "apt",
				"template", "template", "template", "template", "template", "template", "template",
			},
			"ipsum":   {
				"raw", "template",
			},
		},
		{
			"subutai": {
				"raw", "template",
			},
			"lorem":   {
				"raw", "template",
			},
			"ipsum":   {
				"raw", "template",
			},
		},
	}
)


// all user test files directories
var (
	Dirs = []map[string]string{
		// 0-layer stands for public directories
		{
			"subutai": "/tmp/cdn-test-data/public/" + Subutai.Username + "/",
			"lorem":   "/tmp/cdn-test-data/public/" + Lorem.Username + "/",
			"ipsum":   "/tmp/cdn-test-data/public/" + Ipsum.Username + "/",
		},
		// 1-layer stands for private directories
		{
			"subutai": "/tmp/cdn-test-data/private/" + Subutai.Username + "/",
			"lorem":   "/tmp/cdn-test-data/private/" + Lorem.Username + "/",
			"ipsum":   "/tmp/cdn-test-data/private/" + Ipsum.Username + "/",
		},
	}
)

// all user test files' IDs and filenames
var (
	Files = []map[string][][]string{
		// 0-layer stands for public files for download
		{
			// 0-layer stands for IDs
			// 1-layer stands for filenames
			"subutai": {
				{
					"adcafddc-71a4-4c28-87f2-d0cff672c85c",
					"205cfd3d-3a50-458a-96af-f65f1244e27c",
				},
				{
					"generic-ansible-subutai-template_0.4.5_amd64.tar.gz",
					"ubuntu-xenial-subutai-template_0.4.5_amd64.tar.gz",
				},
			},
			"lorem": {
				{
					"f32dd7d5-0c9a-4b1a-bfcf-35909310d6b9", // 0 - 0
					"d674228c-fe0f-4849-b104-467872944058", // 1 - 1
					"542d34f8-5146-49b4-acbe-35953b5f2f7c", // 2 - 2
					"f32dd7d5-0c9a-4b1a-bfcf-35909310d6b9", // 3 - incorrect filename
					"5eab3bc3-1d4c-4252-a066-58ffe98d023e", // 4 - full template config - 3
					"0c531e2f-10d0-4bbb-99be-c7154d33a77d", // 5 - no config
					"8ee172da-5c36-4c9a-9194-3ff295890776", // 6 - borked name & version
					"dae96500-d5d6-4878-8fe0-1a09c416e4e5", // 7 - dependent template - 4
					"98236816-0392-4db8-96a6-4d835180098a", // 8 - semi-empty config
					"be10ba49-bb9c-409d-9182-0a13896b32d8", // 9 - template with bad dependencies
				},
				{
					"generic-ansible-subutai-template_0.4.5_amd64.tar.gz",
					"ubuntu-xenial-subutai-template_0.4.5_amd64.tar.gz",
					"python_2.7.11-1_amd64.deb",
					"lorem-generic-ansible-subutai-template_0.4.5_amd64.tar.gz",
					"generic-ansible-subutai-template_0.4.5_amd64-full.tar.gz",
					"generic-ansible-subutai-template_0.4.5_amd64-empty.tar.gz",
					"generic-ansible^-subutai-template_0.4.5_amd64-name-version.tar.gz",
					"generic-ansible-subutai-template_0.4.7_amd64-dependent.tar.gz",
					"generic-ansible-subutai-template_0.4.5_amd64-semi-empty.tar.gz",
					"generic-ansible-subutai-template_0.4.9_amd64-bad-dependencies.tar.gz",
				},
			},
			"ipsum": {
				{
					"a7471dfa-8c36-40af-82f8-0de1f0ad36c8",
					"93cac025-beba-434e-9dc4-27abeea3e57b",
				},
				{
					"generic-ansible-subutai-template_0.4.5_amd64.tar.gz",
					"ubuntu-xenial-subutai-template_0.4.5_amd64.tar.gz",
				},
			},
		},
		// 1-layer stands for private files for download
		{
			// 0-layer stands for IDs
			// 1-layer stands for filenames
			"subutai": {
				{
					"dab2ed9a-6823-44a3-9b94-e013c89401ce",
					"037e395b-9754-4c41-8c57-ebdd91a68717",
				},
				{
					"generic-ansible-subutai-template_0.4.6_amd64.tar.gz",
					"ubuntu-xenial-subutai-template_0.4.6_amd64.tar.gz",
				},
			},
			"lorem": {
				{
					"2e1807cd-20eb-4afb-bebb-febe5ba9c0d1",
					"aad0c38d-2da7-4b16-9791-44cc37a27f84",
				},
				{
					"generic-ansible-subutai-template_0.4.6_amd64.tar.gz",
					"ubuntu-xenial-subutai-template_0.4.6_amd64.tar.gz",
				},
			},
			"ipsum": {
				{
					"c27443da-593b-4d1f-8d1a-db9e460c8d99",
					"d96926e4-2d8d-4fda-a0e8-3271e2abb797",
				},
				{
					"generic-ansible-subutai-template_0.4.6_amd64.tar.gz",
					"ubuntu-xenial-subutai-template_0.4.6_amd64.tar.gz",
				},
			},
		},
	}
)

var (
	PublicKeys = map[string]string{
		"subutai": SubutaiKey,
		"lorem":   LoremKey,
		"ipsum":   IpsumKey,
	}
)

var (
	SubutaiKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
=R8TO
-----END PGP PUBLIC KEY BLOCK-----
`
)

var (
	LoremKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

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
iTL3bcFog5Zf+w==
=qbAq
-----END PGP PUBLIC KEY BLOCK-----
`
)

var (
	IpsumKey = `-----BEGIN PGP PUBLIC KEY BLOCK-----
Version: GnuPG v1

mQENBFr5Z7wBCACv7AytUuenUbXTs8jd9Pvd3j832vWGHwX9Kf4v6iW4INftGguQ
0Q40nvYKKhFmaeBUNkm3p13iXqq8/rjTFuC+kkDaT2AP72UsQ3KSF6K15cKC8gkF
ktyNZrAraofQ+n4rCGSpHKldf+8bU7+mpnGEEGKsfomOTzWwAsmoilGAseFr5tV9
x5YJIrzcMoiymCqsiOgLTq7vAYPQOl5gcFxlMSt7/cgpWHjTqBCUtoJ6Kvt8mftE
Pd1ycK6mEoQ0qROuazpQCtMeeC5IVFfw27pR1Jrpf6dbMJCqgt9FUIf0aDdrxr+l
TazFSUrTOHYooSP6bDaRjVQ/k5IZnC43M/3lABEBAAG0MGFiYXl0dWxha292YSA8
YWJheXR1bGFrb3ZhQG9wdGltYWwtZHluYW1pY3MuY29tPokBOAQTAQIAIgUCWvln
vAIbAwYLCQgHAwIGFQgCCQoLBBYCAwECHgECF4AACgkQ5L9rtPCyVCeffgf/R0Wz
nIhaji5iwkDHlPR3FyENlWD6KUwj8DhD8t5cTmlPuLVBf7x92TTPJpoqOVtXObgI
yXPPS++49Paubh5F9kIPOca2ilmHReum7XzdR2pAL/EVnkZn37XgW3JN21F645iE
ZmUA/INhpxpXDpRVeF1ZwQ+5/VghXLkZJ4BOwCUwjrtZxb/FhHbCmbrY3Sq1FWZ9
45btKOrhrmlRpB4i3bM+hNZfeqzVpQ1SvboCqJ1r+ZOvUwcZ+RFPUk/z4xiG1d7v
8D3Q66G+6pXmLlAs4/9lQ9wlGMucvChsXIg07qnNkAgaYGh65tI+oEjKDXGPCPr2
bPLPCEwvkIfJivKKybkBDQRa+We8AQgA2AqyFsETnEEOxSbZFphqdRMC8xlEekmj
P9TMZDjrMouRKTJs8Hx/AU9kuKubua0hJWYft8QMlR3JjFkPKhRGs8LkEra7mGSK
6XzKarH5oaF2toFUxIzE6un922wGDt2kq4kiTA0+SI2QX15ZvOX/3+ZYR/q0DV+2
wpsBX2UR3y3UjHIk5BjMU+E9wJNw86098nZ6l8b10xZzwf+rSxpDVwB+1lJTAFbB
ivtoTKxrxmaRKvh909ut6TJA6uU5tGL1AJ2ATysuOLQEmsASnRikxomjt6CMjQGG
Lx3NknQ26BmYfh8rrDQknWw1E0i5mLUgaJJNKv1sRNd4fwhSYAGiWQARAQABiQEf
BBgBAgAJBQJa+We8AhsMAAoJEOS/a7TwslQn3voH/0y0aXSxvtt0gd62+6iA5R6m
Y7bQvf4vyndEXk4+Z3LjmMRTfoxak/AhnxzAkkC2ednVBscZE9btxiP7dkdHlJ4B
lvbR5K7xRJyFgUFUbL+s5GyXh9j0OVruw6qFteX04UD7OTZJxTm/+yIzLwmDfovP
NkfUgyMjtjPjrKWyydoEbofg2WTFgiuA6f/5zHZLG6NNxTgoxz6ZSLhnvUN3JVJ5
YaQ9+dgz72Zum+hRgg7t7cN4VF8iMuK2fYx/grIZ2wH5c4s9wJsPy0D+MiijhaF5
fOBXzFkOhrc9NzsD4RGldnFJQkaYotxOXbPlfosCdc+3WuS0Np+mpnu6/LsfzSQ=
=XRVg
-----END PGP PUBLIC KEY BLOCK-----
`
)
