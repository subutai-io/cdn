use strict;
use warnings;
use Test::More;
use File::Temp qw/ :POSIX /;
use lib qw(./lib);
use constant {
    FSIZE => 4000,
};

BEGIN {
    $ENV{GORJUN_EMAIL}         //= 'tester@gmail.com';
    $ENV{GORJUN_USER}          //= 'tester';
    $ENV{GORJUN_HOST}          //= '127.0.0.1';
    $ENV{GORJUN_PORT}          //= '8080';
    $ENV{MOJO_USERAGENT_DEBUG} //= 0;                    # Set to see requests
}

chomp( my $KEY = `gpg --armor --export $ENV{GORJUN_EMAIL}` );

sub create_file {
    my %args = @_;
    my $size = $args{size} || FSIZE;
    my $fname = tmpnam();

    open( my $fh, '>', $fname );
    print $fh "\0" x (1024*$size);
    close($fh);

    return $fname;
}

use_ok('Gorjun');

my $g = Gorjun->new( gpg_pass_phrase => 'pantano' );

my $test_info = <<EOF;
Test Data Information: 

User Name: %s
User Email: %s
PGP Key:
%s

EOF

note( sprintf $test_info, ( $g->user, $g->email, $g->key ) );

# test registering a user (if not registered already)
ok my $res = $g->register( name => $ENV{GORJUN_USER}, key => $KEY ),
  "Register was done";
note($res);

# test quota checking
ok my $quota = $g->quota( user => $ENV{GORJUN_USER} ), "Get quota value done";
note($quota);

# test get new token
ok my $token = $g->get_token( user => "$ENV{GORJUN_USER}" ), "Token got";
note($token);

# test uploading
my %UPLOADS = (
    raw => { file => create_file(size => 1) },
    template => {
        file =>
          't/Data/abdysamat-apache-subutai-template_4.0.0_amd64.tar.gz',
    },
    apt =>
      { file => 't/Data/winff_1.5.5-1_all.deb' },
);

my @uploads;

while ( my ( $type, $file ) = each %UPLOADS ) {
    my ($res, $upload);
    note("$type : " . $file->{file});

    ok $upload = $g->upload(
        type  => $type,
        file  => $file,
        token => $token
      ), "$type upload done";
    note($upload);

    sleep 3; # some reason gorjun needs time to save to file

    # sign uploaded file
    ok $g->sign( token => $token, signature => $upload ), 'Sign done';

    # list uploaded file in repo
    ok $res = $g->send( method => 'get', path => "/kurjun/rest/$type/info" ),
    "Listing $type repository";
    note($res);

    push @uploads, $upload; # save upload id's
}

# Share repositories
#
# TODO: share repository with another user
#
#for my $type ( qw{ raw apt template }) {
#}


# Delete uploaded files
#
# TODO: test deletion of each uploaded file
#
#foreach my $upload ( @uploads ) {
#
#}

# shutdown server if we are code covering
if ( $ENV{ GORJUN_COVER } ) {
    ok ! eval { $g->send( method => 'get', path => '/kurjun/rest/shutdown' ); },
      "Shutting down gorjun server";
}

# finish test
done_testing() unless $ENV{GORJUN_COVER};
