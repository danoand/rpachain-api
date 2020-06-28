
function configState($stateProvider, $urlRouterProvider, $compileProvider) {

    // Optimize load start with remove binding information inside the DOM element
    $compileProvider.debugInfoEnabled(true);

    // Set default state
    $urlRouterProvider.otherwise("/login");
    $stateProvider

        // Dashboard - Main page
        .state('dashboard', {
            url: "/dashboard",
            templateUrl: "views/dashboard.html",
            data: {
                pageTitle: 'Dashboard',
            }
        })

        // Login Page
        .state('login', {
            url: "/login",
            templateUrl: "views/login.html",
            data: {
                pageTitle: 'Login',
            }
        })

        // App views
        .state('app_views', {
            abstract: true,
            url: "/app_views",
            templateUrl: "views/common/content.html",
            data: {
                pageTitle: 'App Views'
            }
        })
        // (Manually) Add to the Blockchain
        .state('app_views.blockwrite_add', {
            url: "/blockwrite_add",
            templateUrl: "views/app_views/blockwrite_manual_add.html",
            data: {
              pageTitle: 'Add to Chain',
              pageDesc: 'Notarize content and files to the Chain.'
            }
          })
        // View Blockwrite details
        .state('app_views.blockwrite_view', {
            url: "/blockwrite_view/:docid",
            templateUrl: "views/app_views/blockwrite_view.html",
            data: {
              pageTitle: 'View Blockwrite',
              pageDesc: 'View details notarized to the Chain.'
            }
          })
}

angular
    .module('rpachain')
    .config(configState)
    .run(function($rootScope, $state) {
        $rootScope.$state = $state;
    });