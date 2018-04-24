# Nginx Rewrite Unit Test

Nginx Rewrite Unit Test checks the rewrite rules in your Nginx configuration. This tool is a combination 
of 3 applications all running as Docker containers      

* Test Server               
Responsible for parsing the test cases and sending request to the Target Nginx server       

* Target Nginx          
The Nginx server to be tested runs as a docker container.You can either use your
own Nginx image, or just volume mount the nginx configuration.

* Dummy Web Server          
A dummy application, to which all redirected requests from target Nginx will be
sent. It is used for verifying the redirect URL path. It gives as response, the 
host and url path of the request.    

### Prerequisites

* Docker Compose (Version 3)

### Quick Start

You can get started with the Nginx rewrite unit test by cloning this repository and switch to `stand-alone` branch.
This branch is configured with default values, and sample test cases and redirect rules to get a quick overview.

1. Clone the repo       
`git clone git@github.com:fastretailing/rewrite_unit_test.git --branch stand-alone`

2. `cd rewrite_unit_test`

3. `docker-compose up`

### Steps

1. Clone the repo       
    ```
    git clone git@github.com:fastretailing/rewrite_unit_test.git
    ``` 
2. a. Add the test urls to the `test_server\test-urls.yml` file in the YAML format.  
          
    ```
    - name: Test Name
      test: http://host.com/some/path
      expected: http://host.com/redirect/path
    ...
     ```

    b. Export TestCases file path as following:
     
    ```
    export TESTCASES_FILEPATH=./test_server/test-urls.yml        
    ```
   *Note: Filepath should be relative to the `docker-compose.yml` file.*
   
3. Uncomment the `target_nginx` portion in `docker-compose.yml` file. 

    *Note: `target_nginx` service is commented to be more generic. Configure `target_nginx` as you require.*

    * If you want run tests on your Nginx configuration directly, export the nginx image.  
       `export NGINX_IMAGE=nginx`        
    
    OR,
    
    * If you have your own Nginx image, configure the `target_nginx` service in Docker Compose accordingly.    
       `export NGINX_IMAGE=<your_nginx_image>`        
       Configure the environment variables and other configurations required for your nginx application.
   
 
4. *(Skip if you are using your own Nginx image)* Replace the Nginx configuration file with 
`target_nginx/nginx.conf` file, or modify the rewrite rules in the existing file.          

    ```
        server {
            server_name host;
            rewrite ^/some/path /redirect/path permanent;
        }
    ```

    *Note: In Nginx container*
        - `/target_nginx/nginx.conf` is volume mounted with the `/etc/nginx/nginx.conf` file.
        - `/target_nginx/cert.crt` is volume mounted with the `/etc/ssl/certs/cert.crt` file.
        - `/target_nginx/cert.key` is volume mounted with the `/etc/ssl/certs/cert.key` file.

5. Deploy using the Docker Compose     
 
    `docker-compose up --build --force-recreate`

6. Open the Test Server from your browser          
 
    `http://localhost:6582`
   

### As part of Jenkins Pipeline

`unit-test-server` can be added as part of your Jenkins Pipeline as follows:

1. Add the test cases in `test_urls.yml` file in your Nginx repo.

2. Modify your Jenkinsfile so that unit-test will be run after building the Nginx image.

    a. Define the following variables    
    ```
    def nginx_image="<your_nginx_image>"                                 
    def testcases_filepath='./test-urls.yml'
    def static_network_prefix = env.JOB_NAME.replaceAll("[^a-zA-Z0-9]+","")
    ```

    b. Keep the `test-urls.yml` from your nginx repo in the Jenkins Workspace 
    before cloning the `unit-test-server` repo.
    ```
    sh 'mv test_urls.yml test-urls.yml' 
    ```

    c. Add the `unit-test` stage where we will be running test-cases and generating reports.
     
     ```
       stage('unit-test') {
         git url: "git@github.com:fastretailing/rewrite_unit_test.git", credentialsId: "<jenkins_credentials_id>"
     
     
         withEnv([
                 "TESTCASES_FILEPATH=${testcases_filepath}",
                 "NGINX_IMAGE=${nginx_image}",
         ]) {
     
         sh "docker-compose up --build --force-recreate -d"
     
         /* Running your nginx image for unit-test */
         
         /* Add the run options as required by your Nginx image   */
         /* Add 172.20.128.30 as the IP address of any hosts to   */
         /* which the requests will be redirected from your Nginx */
         
         sh "docker run \
         --net ${static_network_prefix}_static-network --ip 172.20.128.20 \
         --add-host example-a:172.20.128.30 \
         --add-host example-b:172.20.128.30 \
         -e foo=bar \
         -d ${nginx_image}"
     
         /* Running unit-tests */
         sh 'curl http://localhost:6582/run-test > report.xml'
     
         /* Cleanup */
         sh 'docker rm -f $(docker ps -lq)'
         sh "docker-compose down"
     
         /* Report generation */
         junit 'report.xml'
         }
       }
     ```
 
    d. (Optional) You can prevent from pushing your failed images to the registry by 
     checking whether the current build is success.
     ```
     if (currentBuild.result == null){
         /* If unit tests have failed `currentBuild` will be 'UNSTABLE' and stage 'push' will not be triggerd */
         stage("push"){
         ..................
         ..................
         }
     ```
 
