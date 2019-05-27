function setPreviewBoxForElements(imageLinks) {
    for (var i = 0; i < imageLinks.length; i++) {
        // imageLinks[i].addEventListener('click', function (e) {
        //     e.preventDefault();
        //     BigPicture({
        //         el: e.target,
        //         gallery: '#photos',
        //         // vidSrc: e.target.getAttribute('data-video-src'),
        //         noloader: false
        //     });
        // })

        if (imageLinks[i].classList.contains('video')) {
            imageLinks[i].addEventListener('click', function (e) {
                e.preventDefault();
                BigPicture({
                    el: e.target,
                    //gallery: '#photos',
                    vidSrc: e.target.getAttribute('data-bp'),
                    noloader: false
                });
            })
        } else {
            imageLinks[i].addEventListener('click', function (e) {
                e.preventDefault();
                BigPicture({
                    el: e.target,
                    gallery: '#photos',
                    noloader: true
                });
            })
        }


    }

    // var videoLinks = document.querySelectorAll('a.video-link');
    // for (var i = 0; i < videoLinks.length; i++) {
    //     videoLinks[i].addEventListener('click', function (e) {
    //         e.preventDefault();
    //         BigPicture({
    //             el: e.target,
    //             // gallery: '#photos',
    //             vidSrc: e.target.getAttribute('data-video-src'),
    //             noloader: false
    //         });
    //     })
    // }
}

document.addEventListener('DOMContentLoaded', function () {

    setPreviewBoxForElements(document.querySelectorAll('a.jsp'));

    // yall({observeChanges: true, idleLoadTimeout: 3000});

    var conf = document.querySelector('#conf');
    if (false && parseInt(conf.dataset.maxpage) > 1) {

        var infScroll = new InfiniteScroll('.photos', {
            // defaults listed

            path: '.next-page',
            // REQUIRED. Determines the URL for the next page
            // Set to selector string to use the href of the next page's link
            // path: '.pagination__next'
            // Or set with {{#}} in place of the page number in the url
            // path: '/blog/page/{{#}}'
            // or set with function
            // path: function() {
            //   return return '/articles/P' + ( ( this.loadCount + 1 ) * 10 );
            // }

            append: '.element',
            // REQUIRED for appending content
            // Appends selected elements from loaded page to the container

            checkLastPage: true,
            // Checks if page has path selector element
            // Set to string if path is not set as selector string:
            //   checkLastPage: '.pagination__next'

            prefill: true,
            // Loads and appends pages on intialization until scroll requirement is met.

            responseType: 'document',
            // Sets the type of response returned by the page request.
            // Set to 'text' to return flat text for loading JSON.

            outlayer: false,
            // Integrates Masonry, Isotope or Packery
            // Appended items will be added to the layout

            scrollThreshold: 400,
            // Sets the distance between the viewport to scroll area
            // for scrollThreshold event to be triggered.

            elementScroll: false,
            // Sets scroller to an element for overflow element scrolling

            loadOnScroll: true,
            // Loads next page when scroll crosses over scrollThreshold

            history: 'replace',
            // Changes the browser history and URL.
            // Set to 'push' to use history.pushState()
            //    to create new history entries for each page change.

            historyTitle: true,
            // Updates the window title. Requires history enabled.

            hideNav: undefined,
            // Hides navigation element

            status: '.page-load-status',
            // Displays status elements indicating state of page loading:
            // .infinite-scroll-request, .infinite-scroll-load, .infinite-scroll-error
            // status: '.page-load-status'

            button: undefined,
            // Enables a button to load pages on click
            // button: '.load-next-button'

            onInit: undefined,
            // called on initialization
            // useful for binding events on init
            // onInit: function() {
            //   this.on( 'append', function() {...})
            // }

            debug: false,
            // Logs events and state changes to the console.
        });

        infScroll.on('append', function (response, path, items) {
            for (var i = 0; i < items.length; i++) {
                setPreviewBoxForElements(items[i].querySelectorAll('a.jsp'));
            }
        });

        infScroll.on('error', function (error, path) {
            console.log("Error: " + error + " on path: " + path);
        });
    }
});