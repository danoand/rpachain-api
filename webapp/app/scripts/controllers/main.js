/**
 *
 * appCtrl
 *
 */

angular
    .module('rpachain')
    .controller('appCtrl',                  appCtrl)
    .controller('loginCtrl',                loginCtrl)
    .controller('hdrCtrl',                  hdrCtrl)
    .controller('navCtrl',                  navCtrl)
    .controller('dashCtrl',                 dashCtrl)
    .controller('dashBlockWritesTableCtrl', dashBlockWritesTableCtrl)
    .controller('writeManualBlockCtrl',     writeManualBlockCtrl)
    .factory('sessSvc',                     sessSvc);

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

// dashBlockWritesTableCtrl controls the the block write table on the dashboard view
function dashBlockWritesTableCtrl($http, $scope, $state, $window, sessSvc) {
    // $scope.myData = [{"name": "Tom"}, {"name": "Harry"}];
    var prms = sessSvc.getUserData()
    var hdrs = {};
    hdrs["X-username"] = prms.username;
    hdrs["X-docid"] = prms.docid;
    var selected_rows = [];

    $scope.blockURL = '';

    $scope.gridOptions = {
        showGridFooter: true,
        enableRowSelection: true,
        multiSelect: false,
        enableRowHeaderSelection: false,
        columnDefs: [
          { name: 'network', enableSorting: true },
          { name: 'timestamp', enableSorting: true },
          { name: 'block', enableSorting: true },
          { name: 'action', enableSorting: false },
          { name: 'explorer_link', enableSorting: false }
        ],
        onRegisterApi: function(gridApi) {
          $scope.gridApi = gridApi;

          gridApi.selection.on.rowSelectionChanged($scope,function(row){
            $scope.blockURL = '';
            selected_rows   = [];

            var tmp_count = gridApi.selection.getSelectedCount();
            if (tmp_count == 1) {
                selected_rows   = gridApi.selection.getSelectedRows();
                $scope.blockURL = selected_rows[0]["explorer_link"];
            }
          });
        }
      };

    // Call backend to validate username and password
    $http({
        method: 'GET',
        url: '/webapp/getblockwrites',
        headers: hdrs
    }).then(function successCallback(response) {
        $scope.gridOptions.data = response.data.content; 
    }, function errorCallback(response) {
        // Authentication was failed
        console.log('ERROR: Error callback for /webapp/getblockwrites with response: ' + JSON.stringify(response));

        growl.warning(response.data.msg, {ttl: 2500});
    });

    // View block explorer block
    $scope.displayBlockExploreBlock = function (to_url) {
        if ($scope.blockURL.length == 0) {
            return
        }
        $window.open($scope.blockURL, '_blank');
    };

    // Display the 'Add to Chain' Form
    $scope.gotoAddBlockWrite = function() {
        $state.go('app_views.blockwrite_add');
    };
}

// writeManualBlockCtrl controls the manual writing of a block
function writeManualBlockCtrl($http, $scope, $state, growl, sessSvc) {
    $scope.config = {};
    $scope.block  = {};

    $scope.config.highlight   = "#ffff99";
    $scope.config.isProcessing       = false;

    $scope.block.title        = "";
    $scope.block.network      = "gochain_testnet";
    $scope.block.content_text = "";
    $scope.block.meta_data_01 = "";

    // Validate validates required
    var validate = function () {
        var errMsg = [];

        // Check for required fields

        // Title
        if ($scope.block.title === undefined || $scope.block.title == "") {
            errMsg.push("title is missing");
        }
        // Title
        if ($scope.block.network === undefined || $scope.block.network == "None") {
            errMsg.push("blockchain network is missing");
        }

        var errMsgTxt = errMsg.join("; ");

        return errMsgTxt;
    };

    // updUndefOptValues makes sure that undefined, optional values are set to an empty string
    var updUndefOptValues = function () {
        if ($scope.block.content_text === undefined) {
            $scope.block.content_text = "";
        }

        if ($scope.block.meta_data_01 === undefined) {
            $scope.block.meta_data_01 = "";
        }

        return;
    }

    // filePost uploads a file and additional data to the backend
    var filePost = function () {
        var file = $scope.myFile;
        var fd = new FormData();

        var prms = sessSvc.getUserData()
        var hdrs = {};
        hdrs["X-username"]      = prms.username;
        hdrs["X-docid"]         = prms.docid;
        hdrs["Content-Type"]    = undefined;

        $scope.config.isProcessing = true;

        // Update undefined, optional values
        updUndefOptValues();

        fd.append('title', $scope.block.title.toString());
        fd.append('network', $scope.block.network.toString());
        fd.append('content_text', $scope.block.content_text.toString());
        fd.append('meta_data_01', $scope.block.meta_data_01.toString());
        fd.append('file', file);

        // Post the data to the backend
        $http.post('/webapp/manualblockwrite/upload', fd, {
            transformRequest: angular.identity,
            headers: hdrs
        }).success(function (response) {
            growl.success(response.data.msg, {ttl: 2500});
            $scope.config.isProcessing = false;
            $state.go('dashboard');
        }).error(function (response) {
            // post of a new product failed
            console.log('ERROR: error callback for /updategeneralpromotion with response: ' + JSON.stringify(response));
            $scope.config.isProcessing = false;
            growl.danger(response.data.msg, {ttl: 2500});
        });
    };

    // prmsPost posts data (not a file to the backend)
    var prmsPost = function () {
        // Define the parameter inbound to the server
        var inbndData = {};

        $scope.config.isProcessing = true;

        // Update undefined, optional values
        updUndefOptValues();

        inbndData.title         = $scope.block.title.toString();
        inbndData.network       = $scope.block.network.toString();
        inbndData.content_text  = $scope.block.content_text.toString();
        inbndData.meta_data_01  = $scope.block.meta_data_01.toString();

        $http({
            method: 'POST',
            url: '/webapp/manualblockwrite/noupload',
            data: inbndData
        }).then(function successCallback(response) {
            // post of a new product was success
            $scope.config.isProcessing = false;

            // Message the user
            growl.success(response.data.msg, {ttl: 2500});

            $state.go('dashboard');
        }, function errorCallback(response) {
            // post of a new product failed
            console.log('ERROR: error callback for updategeneralpromotion with response: ' + JSON.stringify(response));

            // Message the user
            growl.error(response.data.msg, {ttl: 2500});

            $scope.config.isProcessing = false;
        });
        return;
    };

    $scope.save = function () {
        // Validate the form values
        var valErr = validate();
        console.log('DEBUG: just after the validate function with errors: ' + valErr);
        if (valErr.length > 0) {
            // Validation errors exist, notify the user
            growl.warning('Some errors are present: ' + valErr, {ttl: 2500});

            return;
        }

        // Message the user
        growl.info('Your content is being notarized.', {ttl: 2500});

        // Determine if the user has selected an upload file
        console.log('DEBUG: type of $scope.myFile is: ' + typeof $scope.myFile);
        var myFile_type = typeof $scope.myFile
        if (myFile_type == "undefined") {
            console.log('DEBUG: just inside the if statement');
            // No upload file selected - just post the form details (but no file upload)
            prmsPost();
            return;
        }

        // User has selected an upload file
        console.log('DEBUG: just before the filePost function call');
        filePost();
        return;
    };
}

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

