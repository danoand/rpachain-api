/**
 *
 * appCtrl
 *
 */

angular
    .module('rpachain')
    .controller('appCtrl',      appCtrl)
    .controller('loginCtrl',    loginCtrl)
    .controller('hdrCtrl',      hdrCtrl)
    .controller('navCtrl',      navCtrl)
    .controller('dashCtrl',     dashCtrl)
    .factory('sessSvc',         sessSvc);

function appCtrl($http, $scope) {};

// loginCtrl controls the Login view
function loginCtrl($http, $scope, $state, $cookies, growl, sessSvc) {
    $cookies.remove('go_session_id');

    // Set initial values
    $scope.username = "";
    $scope.password = "";
    
    // Submit the login form
    $scope.submit = function () {

        // Call backend to validate username and password
        $http({
            method: 'POST',
            url: '/webapp/login',
            data: {
                username: $scope.username,
                password: $scope.password
            }
        }).then(function successCallback(response) {
            console.log(JSON.stringify(response));
            sessSvc.setUserData(response.data.content);
            growl.success(response.data.msg, {ttl: 1000});
            sessSvc.dumpUserData();

            $state.go('dashboard');
            
        }, function errorCallback(response) {
            // Authentication was failed
            console.log('ERROR: Error callback for /login with response: ' + JSON.stringify(response));

            growl.warning(response.data.msg, {ttl: 2500});
        });
    };
};

// hdrCtrl controls the common > header view
function hdrCtrl($http, $scope, $state, growl, $cookies) {
    
    // Logoff of the web application
    $scope.logOff = function () {
        $cookies.remove('go_session_id');

        // Call backend to validate username and password
        $http({
            method: 'POST',
            url: '/webapp/logoff'
        }).then(function successCallback(response) {
            console.log(JSON.stringify(response));

            growl.success(response.data.msg, {ttl: 1000});
            $state.go('login');
            
        }, function errorCallback(response) {
            // Authentication was failed
            console.log('ERROR: Error callback for /login with response: ' + JSON.stringify(response));
            growl.warning(response.data.msg, {ttl: 2500});
        });
    };
};

// navCtrl controls the common > navigation view
function navCtrl($scope, sessSvc) {

    // Get the username from the user's session
    var sess = sessSvc.getUserData();
    $scope.username = sess["username"] || "Hello!";
    
    // Logoff of the web application
    $scope.logOff = function () {
        $cookies.remove('go_session_id');

        // Call backend to validate username and password
        $http({
            method: 'POST',
            url: '/webapp/logoff'
        }).then(function successCallback(response) {
            console.log(JSON.stringify(response));

            growl.success(response.data.msg, {ttl: 1000});
            $state.go('login');
            
        }, function errorCallback(response) {
            // Authentication was failed
            console.log('ERROR: Error callback for /login with response: ' + JSON.stringify(response));
            growl.warning(response.data.msg, {ttl: 2500});
        });
    };
};

// dashCtrl controls the common > dashboard view
function dashCtrl($scope, sessSvc) {

    // Get the username from the user's session
    var sess = sessSvc.getUserData();
    $scope.accountname = sess["accountname"] || "Hello!";

    // Set up chart
    $scope.labels = [
        "March 27", 
        "April 6", 
        "April 13", 
        "April 20", 
        "April 27", 
        "May 4", 
        "May 11"];
    $scope.series = ['Block Updates', 'Contract Calls'];
    $scope.data = [
      [65, 59, 80, 81, 56, 55, 40],
      [28, 48, 40, 19, 86, 27, 90]
    ];
    $scope.onClick = function (points, evt) {
      console.log(points, evt);
    };
    $scope.datasetOverride = [{ yAxisID: 'y-axis-1' }, { yAxisID: 'y-axis-2' }];
    $scope.options = {
      scales: {
        yAxes: [
          {
            id: 'y-axis-1',
            type: 'linear',
            display: true,
            position: 'left'
          },
          {
            id: 'y-axis-2',
            type: 'linear',
            display: true,
            position: 'right'
          }
        ]
      }
    };
    
};

// sessSvc provides user session type services
function sessSvc() {
    // Define variables that house data for this service
    var sess_user = {};
    sess_user.docid         = "";
    sess_user.username      = "";
    sess_user.accountname   = "";

    // Define methods associated with this service
    return {
        // Set user session data
        setUserData: function(inval) {
            sess_user.docid         = inval["docid"];
            sess_user.username      = inval["username"];
            sess_user.accountname   = inval["accountname"];
        },
        getUserData: function() {
            return sess_user;
        },
        dumpUserData: function() {
            console.log('DEBUG: user session data is: ' + JSON.stringify(sess_user));
        }
    };
};

