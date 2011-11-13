$(function () {
    var d2 = [[0, 3], [4, 8], [8, 5], [9, 13]];

    $.plot($('#chart'), [
        {
            data: d2,
            bars: { show: true }
        }
    ]);
});
