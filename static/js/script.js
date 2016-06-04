$(function(){
	$('.navIcon').click(function(){
		$('.overlay').toggle();
		$('.menu').toggleClass('menuOn');
		$('#wrap').toggleClass('fixed');

		if ($('.iconLayer').hasClass('arrow')) {
			$('.iconLayer').removeClass('arrow').addClass('hamburger');
		} else {
			$('.iconLayer').removeClass('hamburger').addClass("arrow");
		}
		return false;
	});

	$('#contents').prepend('<div class="overlay"></div>');

	$('.overlay').click(function(){
		$(this).fadeOut(300);
		$('.menu').removeClass('menuOn');
		$('#wrap').removeClass('fixed');
		$('.iconLayer').removeClass('arrow').addClass('hamburger');
	});

	function winHeight() {
		var winH = $(window).height();
		var headerH = $('header').height() + 1;
		var winH = winH - headerH;
		$('.menu').css({
			'height': winH + 'px',
			'top': headerH + 'px'
		});
	}
	winHeight();

	$(window).resize(function(){
		winHeight();
	});

	$('.menuHeader figcaption').click(function(){
		$('.menu .txt').slideToggle();
	});

});
