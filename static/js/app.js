document.addEventListener('DOMContentLoaded', function () {

    var imageLinks = document.querySelectorAll('a.jsp');
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
});