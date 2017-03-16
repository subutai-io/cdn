#!groovy

notifyBuildDetails = ""
cdnHost = ""

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
		/* Build subutai binary */
		notifyBuildDetails = "\nFailed on Stage - Build gorjun"

		sh """
			export GOPATH=${workspace}/${goenvDir}
			export GOBIN=${workspace}/${goenvDir}/bin
			go get
			go build
		"""

		stage("Deploy Gorjun")
		notifyBuildDetails = "\nFailed on Stage - Deploy Gorjun"

		// switch (env.BRANCH_NAME) {
		// 	case ~/master/: cdnHost = "stagecdn.subut.ai"; break;
		// 	default: cdnHost = "devcdn.subut.ai"
		// }

		// test node
		cdnHost = "54.194.92.172"

		sh """
			set +x
			scp gorjun root@${cdnHost}:/tmp
			ssh root@${cdnHost} <<- EOF
			set -e

			/bin/mv /tmp/gorjun /mnt/lib/lxc/gorjun/opt/gorjun/bin/gorjun
			/apps/bin/lxc-attach -n gorjun -- systemctrl restart gorjun
		EOF"""	
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
  def slackToken = getSlackToken('sysnet-bots-slack-token')
  // Send notifications
  // slackSend (color: colorCode, message: summary, teamDomain: 'subutai-io', token: "${slackToken}")
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