var root = $($('h1 a').first()).attr('href');

function element(name, content, attributes) {
    var e = $(document.createElement(name));

    if (content) { e.append(content); }
    for (a in attributes) { e.attr(a, attributes[a]); }

    return e;
}

// http://stackoverflow.com/questions/2901102/2901298#2901298
function thousands(x) {
    return x.toString().replace(/\B(?=(?:\d{3})+(?!\d))/g, ',');
}

function bookInfoItemFunction(elem, parent) {
    return function(title, item, func) {
        var definition = elem.data(item);

        if (func) { definition = func(elem); }
        if (definition == '') { return false; }

        var item = [element('dt', title), element('dd', definition)];

        $().append.apply(parent, item);

        return item;
    };
}

function stars(elem) {
    var rating = parseInt(elem.data('rating'));

    if (rating == 0) { return false; }

    return element('img', null, {
        src: root + 'ext/star-' + rating + '.gif',
        width: 16 * rating,
        height: '16',
        alt: rating + ' stars'
    });
}

function averageRating(elem) {
    return [
        elem.data('average'),
        ' (',
        thousands(parseInt(elem.data('ratings'))),
        ')'
    ].join('');
}

$(function() {
    var userName = $('#user-name').text();

    $('td div.reviews a').each(function(i) {
        var bookInfo = element('dl');
        var bookInfoItem = bookInfoItemFunction($(this), bookInfo);

        $.each(['Author', 'Published', 'Pages'], function(i, n) {
            bookInfoItem(n, n.toLowerCase());
        });

        bookInfoItem(userName + "&#8217;s rating", '', stars);
        bookInfoItem('Overall rating', '', averageRating);

        $(this).qtip({
            // Only pre-render the first 100 tooltips.
            prerender: i < 100,
            content: {
                text: bookInfo,
                title: $(this).data('title')
            }
        });
    });
});
