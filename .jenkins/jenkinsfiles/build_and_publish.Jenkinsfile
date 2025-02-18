def major = ''
def minor = ''
def patch = ''
def tag = ''
def dockerImage = ''

pipeline {
  agent { label 'linux' }

  environment {
    // Credential ID from https://jenkins.service.internal.projecticeland.net/credentials/
    DOCKER_REGISTRY = "https://docker-virtual.artifactory.service.internal.projecticeland.net"

    apiRegistry = 'omniex/quickfix'
  }

  stages {

    stage('Build Server') {
      steps {
        script {
          withCredentials([sshUserPrivateKey(credentialsId: 'github-molly-brown-ssh-key', keyFileVariable: 'GITHUB_KEY')]) {
            withEnv(["GIT_SSH_COMMAND=ssh -i $GITHUB_KEY"]) {
              docker.withRegistry(DOCKER_REGISTRY, "artifactory-m-devprod") {
                dockerImage = docker.build("$apiRegistry:latest", "--pull -f ./Dockerfile .")
              }
            }
          }
        }
      }
    }

    stage('Get Current Version') {
      steps {            
        script {
           withCredentials([sshUserPrivateKey(credentialsId: 'github-molly-brown-ssh-key', keyFileVariable: 'GITHUB_KEY')]) {
            withEnv(["GIT_SSH_COMMAND=ssh -i $GITHUB_KEY"]) {
              // find the latest tag from remote, default to 1.0.0 if it doesn't exist
              sh "git remote set-url origin git@github.com:omni3x/quickfix.git"
              def command = $/ git ls-remote --quiet --tags --refs | awk -v def=1.0.0 -F\\\/ '{ print $3 } END { if(NR==0) {print def} }' | sort -V | tail -n 1/$
              def version = sh(returnStdout: true, script: command).trim()

              echo "found version ${version}"
              echo "Current Branch ${env.GIT_BRANCH}"

              def versions = version.split('\\.')
              major = versions[0]
              minor = versions[1]
              patch = versions[2]
            }
          } 
        }
      }
    }

    stage('Set New Version') {
      when {
        expression {
          return ("${env.GIT_BRANCH}" == 'master' || "${env.GIT_BRANCH}" == 'release' || "${env.GIT_BRANCH}" == 'origin/master' || "${env.GIT_BRANCH}" == 'origin/release' )
        }
      }

      steps {
        script {
          if ("${env.GIT_BRANCH}" == 'master' || "${env.GIT_BRANCH}" == 'origin/master' ) {
            def new_patch = patch as Integer
            new_patch++
            tag = "${major}.${minor}.${new_patch}"
            echo "new version ${tag}" 
            slackSend(
                    channel: '#eotc-jenkins',
                    message: "quickfix tag ${tag} was created",
                    color: 'good',
                    teamDomain: 'iceland',
                    tokenCredentialId: '2b40808d-04b7-415b-8401-a4d500a84aab'
            )
          } else { // release branch
            def new_minor = minor as Integer
            new_minor++
            tag = "${major}.${new_minor}.0"
            echo "new version ${tag}"
            slackSend(
                    channel: '#eotc-jenkins',
                    message: "quickfix tag ${tag} was created",
                    color: 'good',
                    teamDomain: 'iceland',
                    tokenCredentialId: '2b40808d-04b7-415b-8401-a4d500a84aab'
            )
          }
        }
      }
    }

    //push tag to GitHub
    stage('Publish Tag') {
      when {
        expression {
          return tag != ''
        }
      }

      steps {
        script {
          // Push to Github
           withCredentials([sshUserPrivateKey(credentialsId: 'github-molly-brown-ssh-key', keyFileVariable: 'GITHUB_KEY')]) {
            withEnv(["GIT_SSH_COMMAND=ssh -i $GITHUB_KEY"]) {
              sh("""
                git checkout origin/${branch_name}
                git pull origin ${branch_name}
                git tag -f -a ${tag} -m \"[Jenkins CI] ${tag}\"
                git push -f origin ${tag}
              """)
            }
          }
        }
      }
    }
  }

  post {
    success {
      slackSend(
        channel: '#eotc-jenkins',
        message: "SUCCESS: quickfix was released. version: ${tag}. '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})",
        color: 'good',
        teamDomain: 'iceland',
        tokenCredentialId: '2b40808d-04b7-415b-8401-a4d500a84aab'
      )
    }
    failure {
      slackSend(
        channel: '#eotc-jenkins',
        message: "FAILED: quickfix failed to be released. version: ${tag}. '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})",
        color: 'danger',
        teamDomain: 'iceland',
        tokenCredentialId: '2b40808d-04b7-415b-8401-a4d500a84aab'
      )
    }
  }  
}