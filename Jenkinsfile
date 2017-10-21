#!groovy

notifyBuildDetails = ""
cdnHost = ""
version = ""

try {
	notifyBuild('STARTED')

	/* Building agent binary.
	Node block used to separate agent and subos code.
	*/
	node() {
		String goenvDir = ".goenv"
		deleteDir()

		stage("Checkout source")
		/* checkout agent repo */
		notifyBuildDetails = "\nFailed on Stage - Checkout source"

		checkout scm

		stage("Prepare GOENV")
		/* Creating GOENV path
		Recreating GOENV path to catch possible issues with external libraries.
		*/
		notifyBuildDetails = "\nFailed on Stage - Prepare GOENV"

		sh """
			if test ! -d ${goenvDir}; then mkdir -p ${goenvDir}/src/github.com/subutai-io/; fi
			ln -s ${workspace} ${workspace}/${goenvDir}/src/github.com/subutai-io/gorjun
		"""

		stage("Build gorjun")
		/* Build gorjun binary */
		notifyBuildDetails = "\nFailed on Stage - Build gorjun"

		version = sh (script: """
			export GOPATH=${workspace}/${goenvDir}
			export GOBIN=${workspace}/${goenvDir}/bin
			export GIT_BRANCH=${env.BRANCH_NAME}
			export PATH=\$PATH:/usr/local/go/bin
			go get
			make | cut -d ' ' -f 7 | cut -d '=' -f 2 | tr -d '\"'
			""", returnStdout: true).trim()

		stage("Deploy Gorjun")
		/* Build binary 
		** upload it to appropriate cdn server 
		*/
		notifyBuildDetails = "\nFailed on Stage - Deploy Gorjun"

		switch (env.BRANCH_NAME) {
			case ~/master/: cdnHost = "stagecdn.subut.ai"; break;
			default: cdnHost = "devcdn.subut.ai"
		}

		sh """
			set +x
			scp -P 8022 gorjun root@${cdnHost}:/tmp
			ssh root@${cdnHost} -p8022 <<- EOF
			set -e
			/bin/mv /tmp/gorjun /var/snap/subutai-dev/common/lxc/gorjun/opt/gorjun/bin/
			sudo subutai attach gorjun systemctl restart gorjun
		EOF"""

		// check remote gorjun version
		sh """
			[ "${version}" == "\$(ssh root@${cdnHost} -p8022 sudo subutai attach gorjun curl -s -q http://127.0.0.1:8080/kurjun/rest/about)" ]
		"""
	}

} catch (e) { 
	currentBuild.result = "FAILED"
	throw e
} finally {
	// Success or failure, always send notifications
	notifyBuild(currentBuild.result, notifyBuildDetails)
}

// https://jenkins.io/blog/2016/07/18/pipline-notifications/
def notifyBuild(String buildStatus = 'STARTED', String details = '') {
  // build status of null means successful
  buildStatus = buildStatus ?: 'SUCCESSFUL'

  // Default values
  def colorName = 'RED'
  def colorCode = '#FF0000'
  def subject = "${buildStatus}: Job '${env.JOB_NAME} [${env.BUILD_NUMBER}]'"  	
  def summary = "${subject} (${env.BUILD_URL})"

  // Override default values based on build status
  if (buildStatus == 'STARTED') {
    color = 'YELLOW'
    colorCode = '#FFFF00'  
  } else if (buildStatus == 'SUCCESSFUL') {
    color = 'GREEN'
    colorCode = '#00FF00'
  } else {
    color = 'RED'
    colorCode = '#FF0000'
	summary = "${subject} (${env.BUILD_URL})${details}"
  }
  // Get token
  def slackToken = getSlackToken('sysnet')
  // Send notifications
  slackSend (color: colorCode, message: summary, teamDomain: 'optdyn', token: "${slackToken}")
}

// get slack token from global jenkins credentials store
@NonCPS
def getSlackToken(String slackCredentialsId){
	// id is ID of creadentials
	def jenkins_creds = Jenkins.instance.getExtensionList('com.cloudbees.plugins.credentials.SystemCredentialsProvider')[0]

	String found_slack_token = jenkins_creds.getStore().getDomains().findResult { domain ->
	  jenkins_creds.getCredentials(domain).findResult { credential ->
	    if(slackCredentialsId.equals(credential.id)) {
	      credential.getSecret()
	    }
	  }
	}
	return found_slack_token
}
