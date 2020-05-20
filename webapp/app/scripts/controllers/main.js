/**
 *
 * appCtrl
 *
 */

angular
    .module('rpachain')
    .controller('appCtrl', appCtrl)
    .controller('loginCtrl', loginCtrl)
    .controller('navCtrl', navCtrl)
    .factory('sessSvc', sessSvc);

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

// navCtrl controls the common > navigation view
function navCtrl($http, $scope, $state, growl, $cookies) {
    
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

// sessSvc provides user session type services
function sessSvc() {
    // Define variables that house data for this service
    var sess_user = {};
    sess_user.docid     = "";
    sess_user.username  = "";

    // Define methods associated with this service
    return {
        // Set user session data
        setUserData: function(inval) {
            sess_user.docid     = inval["docid"];
            sess_user.username  = inval["username"];
        },
        getUserData: function() {
            return sess_user;
        },
        dumpUserData: function() {
            console.log('DEBUG: user session data is: ' + JSON.stringify(sess_user));
        }
    };
};

