var monthNames = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul',
                  'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

var bookFields = ['date', 'title', 'author', 'pages', 'rating',
                  'published'];

// https://developer.mozilla.org/en/JavaScript/Reference/Global_Objects/Array/indexOf#Compatibility
if (!Array.prototype.indexOf) {
    Array.prototype.indexOf = function (searchElement /*, fromIndex */ ) {
        "use strict";
        if (this === void 0 || this === null) {
            throw new TypeError();
        }
        var t = Object(this);
        var len = t.length >>> 0;
        if (len === 0) {
            return -1;
        }
        var n = 0;
        if (arguments.length > 0) {
            n = Number(arguments[1]);
            if (n !== n) { // shortcut for verifying if it's NaN
                n = 0;
            } else if (n !== 0 && n !== Infinity && n !== -Infinity) {
                n = (n > 0 || -1) * Math.floor(Math.abs(n));
            }
        }
        if (n >= len) {
            return -1;
        }
        var k = n >= 0 ? n : Math.max(len - Math.abs(n), 0);
        for (; k < len; k++) {
            if (k in t && t[k] === searchElement) {
                return k;
            }
        }
        return -1;
    }
}

// Converts the Goodreads date format to a JavaScript date object. The
// format is, in Ruby terms:
//   %a %b %d %H:%M:%S %Z %Y
// For example:
// Thu Dec 30 00:00:00 -0800 2010
function toDate(str) {
    dateParts = str.split(' ');

    return new Date(Date.UTC(dateParts[5],
                             monthNames.indexOf(dateParts[1]),
                             dateParts[2]));
}

function loadBookData() {
    return $('td div.reviews a').map(function(i, link) {
        var e = $(link);
        var data = { url: e.attr('href') };

        $.each(bookFields, function(j, v) { data[v] = e.data(v); });

        return [[toDate(e.data('date')), data]];
    });
}

function labelFor(format) {
    return function(p) { return $.plot.formatDate(p[0], format); }
}

function chart(data, xAxis, yAxis) {
    data = (typeof data == 'undefined') ? bookData : data;
    xAxis = (typeof xAxis == 'undefined') ? 'month' : xAxis;
    yAxis = (typeof yAxis == 'undefined') ? 'pages' : xAxis;

    var chartTitle = [
        (xAxis == 'count') ? 'Books' : 'Pages',
        ' read by ',
        (xAxis == 'month') ? 'month' : 'day'
    ].join('');

    var timeFormat = (xAxis == 'month') ? '%y/%0m' : '%y/%0m/%0d'
    var label = labelFor(timeFormat);
    var tickedOff = [];
    var offsets = {};

    var ticks = $.map(data, function(point, i) {
        var tickLabel = label(point);

        if (tickedOff.indexOf(tickLabel) > -1) { return null; }

        tickedOff.push(tickLabel);

        return [[tickedOff.length - 1, tickLabel]];
    });

    var chartData = $.map(ticks, function(tick, i) {
        var y = 0;

        var points = $.grep(data, function(point) {
            return (label(point) == tick[1]);
        });

        var ys = $.each(points, function(j, point) {
            y += (yAxis == 'count') ?
                1 :
                parseInt('0' + point[1][yAxis], 10);
        });

        return [[i, y]];
    });

    var chartOptions = {
        xaxis: { ticks: ticks, tickLength: 0 },
        yaxis: { tickFormatter: thousands, tickLength: 0 },
        grid: { clickable: true }
    };

    $('#charts h2').text(chartTitle);

    $.plot($('#chart'), [
        {
            data: chartData,
            bars: { show: true, align: 'center' }
        }
    ], chartOptions);

    $('#chart').bind('plotclick', function(event, pos, item) {
        if (item) {
            monthID = ticks[item.dataIndex][1].replace('/', '-');
            location.href = '#calendar-' + monthID;
        }
    });
};

var bookData = loadBookData();

$(function() { chart(); });
