/**
 *
 * appCtrl
 *
 */

angular
    .module('homer')
    .controller('appCtrl', appCtrl)
    .controller('loginCtrl', loginCtrl);

function appCtrl($http, $scope) {};

function loginCtrl($http, $scope) {
    console.log('DEBUG: just inside loginCtrl');

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
            console.log(JSON.stringify(response.data.msg));
            
        }, function errorCallback(response) {
            // Authentication was failed
            console.log('ERROR: Error callback for /login with response: ' + JSON.stringify(response));

        });
    };
};

