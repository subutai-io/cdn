package Gorjun;

use Moose;
use Mojo::UserAgent;
use GnuPG::Interface;
use GnuPG::Handles;
use IO::Handle;
use IO::All;
use List::Util qw(any);
use Carp;

my $DEBUG = 0;

my %ACTIONS = (
    register => {
        method    => 'post',
        path      => '/kurjun/rest/auth/register',
        has_param => 1,
        params    => [qw ( name key )],
    },

    token => {
        method    => 'post',
        path      => '/kurjun/rest/auth/token',
        has_param => 1,
        params    => [qw( message user )],
    },

    upload => {
        method    => 'post',
        path      => '/kurjun/rest/(:type)/upload',
        has_param => 1,
        params    => [qw( file token )],
    },

    info => {
        method    => 'get',
        path      => '/kurjun/rest/(:type)/info',
        has_param => 1,
        params    => [qw( template apt raw )],
    },

    download => {
        method    => 'get',
        path      => '/kurjun/rest/(:type)/download',
        has_param => 1,
        params    => [qw( id name token)],
    },

    authid => {
        method    => 'get',
        path      => '/kurjun/rest/auth/token?user=(:user)',
        has_param => 1,
        params    => [qw(user)],
    },

    quota => {
        method    => 'get',
        path      => '/kurjun/rest/quota?',
        has_param => 1,
        params    => [qw(user token fix)],
    },

    sign => {
        method    => 'post',
        path      => '/kurjun/rest/auth/sign',
        has_param => 1,
        params    => [qw(signature token)],
    }
);

has host => (
    is       => 'ro',
    isa      => 'Str',
    required => 1,
    default  => $ENV{GORJUN_HOST}
);

has port => (
    is       => 'ro',
    isa      => 'Int',
    required => 1,
    default  => $ENV{GORJUN_PORT}
);

has user => (
    is       => 'rw',
    isa      => 'Str',
    required => 1,
    default  => $ENV{GORJUN_USER}
);

has email => (
    is       => 'rw',
    isa      => 'Str',
    required => 1,
    default  => $ENV{GORJUN_EMAIL}
);

has gpg_pass_phrase => (
    is       => 'rw',
    isa      => 'Str',
    required => 1,
    default  => $ENV{GPG_PASS}
);

has ua => (
    is      => 'ro',
    isa     => 'Mojo::UserAgent',
    default => sub { Mojo::UserAgent->new }
);

has key => (
    is      => 'ro',
    isa     => 'Str',
    lazy    => 1,
    builder => '_build_key',
);

has _gpg => (
    is      => 'ro',
    isa     => 'GnuPG::Interface',
    lazy    => 1,
    builder => '_build_gpg',
);

has token => (
    is        => 'rw',
    isa       => 'Str',
    predicate => 'has_token',
);

sub _build_key {
    my $self = shift;
    my ( $name, $email, $pass );

    # user has a key already: use it
    $email = $self->email;
    chomp( my $key = `gpg --armor --export $email` );
    return $key if $key && $key ne "";

    # don't have a key: create a key
    ( $name, $pass ) = ( $self->user, $self->gpg_pass_phrase );
    my $instructions = <<EOI;
     %echo Generating a default key
     Key-Type: default
     Subkey-Type: default
     Name-Real: $name
     Name-Comment: with stupid passphrase
     Name-Email: $email
     Expire-Date: 0
     Passphrase: $pass
     %commit
     %echo done
EOI

    # save instructions to create a key
    my $temp_file = '/tmp/instructions';
    $instructions > io($temp_file);

    # generate and delete temp file
    `gpg --batch --generate-key $temp_file`;
    unlink $temp_file;

    # return generated key
    chomp( $key = `gpg --armor --export $self->email` );
    return $key;
}

sub register {
    my $self   = shift;
    my %params = @_;

    carp "Register in progress:" if $DEBUG;

    my $info = $ACTIONS{'register'};
    $params{key} = $self->key unless $params{key};

    $self->send(
        method => $info->{method},
        path   => $info->{path},
        form   => \%params
    );

}

sub get_authid {
    my $self   = shift;
    my %params = @_;
    my $info   = $ACTIONS{'authid'};

    # change user in url and delete this in params
    $info->{path} =~ s/\(:user\)/$params{user}/mx;
    delete $params{user};

    my $res = $self->send(
        method => $info->{method},
        path   => $info->{path},
    );

    return $res;
}

sub get_token {
    my $self   = shift;
    my %params = @_;

    carp "Token in progress:" if $DEBUG;
    return $self->token if $self->has_token;

    my $info = $ACTIONS{'token'};

    # execute get_authid first to get the author id from gorjun
    my $authid = $self->get_authid( user => $params{user} );
    croak "Could not get authid" unless $authid;

    # sign message with user key, passing pass phrase
    $params{message} = $self->clearsign_msg($authid);

    # send to gorjun
    my $res = $self->send(
        method => $info->{method},
        path   => $info->{path},
        form   => \%params,
    );

    # save token
    $self->token($res);

    return $res;
}

sub has_user {
    my $self = shift;
    my $user = shift;

    return $self->get_authid( user => $user ) ? 1 : 0;
}

sub quota {
    my $self   = shift;
    my %params = @_;
    my $info   = $ACTIONS{'quota'};

    # change path for user
    croak "Quota needs a user" unless $params{user};
    $params{token} = $self->get_token( user => $params{'user'} );
    $params{fix} = 'empty';
    $info->{path} .= join '&', map { "$_=$params{$_}" }
      keys %params;

    my $res = $self->send(
        method => $info->{method},
        path   => $info->{path},
    );

    return $res;
}

sub sign {
    my $self   = shift;
    my %params = @_;

    # check mandatory params
    croak "Needs signature and token"
      unless $params{signature} or $params{token};

    my $info = $ACTIONS{'sign'};

    # gpg signs upload response
    $params{signature} = $self->clearsign_msg( $params{signature} );

    # send sign post
    my $res = $self->send(
        method => $info->{method},
        path   => $info->{path},
        form   => \%params
    );

    return $res;
}

sub upload {
    my $self   = shift;
    my %params = @_;
    my $info   = $ACTIONS{'upload'};

    carp "Upload in progress" if $DEBUG;

    croak "Upload needs a type: raw | apt | template" unless $params{type};
    croak "Upload requires token" unless $params{token};

    # change placeholder in path and delete type from params
    $info->{path} =~ s/\(:type\)/$params{type}/mx;
    delete $params{type};

    # token goes to http header
    my $header = { token => $params{token} };
    delete $params{token};

    my $res = $self->send(
        method => $info->{method},
        path   => $info->{path},
        form   => \%params,
        header => $header
    );

    return $res;
}

sub send {
    my $self = shift;
    my %args = @_;

    my $url     = $self->base_url . $args{'path'};
    my $method  = $args{'method'};
    my $form    = $args{'form'};

    # Use any other headers passed
    my $h       = $args{'header'} || {};
    my $headers = { 'Content-Type' => 'multipart/form-data', %$h };

    my $tx =
        $form
      ? $self->ua->$method( $url => $headers => form => $form )
      : $self->ua->$method($url);

    if ( $tx->success ) {
        return $tx->res->body;
    }
    else {
        my $err = $tx->error;
        my $msg = "Error sending request to " . $tx->req->url;
        $msg   .= " Got this: $err->{message}\n";
        croak $msg;
        return;
    }
}

sub base_url {
    my $self = shift;
    return $self->host . ':' . $self->port;
}

#TODO: fake slow connection
sub send_slow {
    my $self = shift;
    my %args = @_;

    my $url     = $self->base_url . $args{'path'};
    my $method  = $args{'method'};
    my $form    = $args{'form'};

    my $tx_type = $method eq 'get' ? 'GET' : 'POST';
    my $tx = $self->ua->build_tx( $tx_type => $url );

    $tx->req->headers->header( 'Content-Type' => 'multipart/form-data' )
      if $tx_type eq 'POST';

    my ($body, $len);
    if ( $method eq 'POST' ) {
        #TODO: set body data
        $body = $form;

        #TODO: set length of body
        $len += length($_) for keys %$form;
        $len += length($_) for values %$form;
    }
    else { 
        $body = 'Some body here';
        $len  = length $body;
    } 

    # total time to expend, giving rate per unit to sleep
    my $rate = $args{'total_time'} / $len;

    $tx->req->headers->content_length( $len );

    # the way we chunk read body to transfer
    my $drain = sub {
        my $content = shift;

        my $chunk = substr $body, 0, '';
        sleep $rate;
        my $drain = undef unless length $body;
        $content->write( $chunk, $drain );
    };

    $tx->req->content->$drain;
    $tx = $self->ua->start($tx);

    if ( $tx->success ) {
        return $tx->res->body;
    }
    else {
        my $err = $tx->error;
        croak "Error sending request, got this: $err->{message}";
        return;
    }
}

sub _build_gpg {
    my $self = shift;

    # init interface to gpg
    my $gnupg = GnuPG::Interface->new();
    $gnupg->options->hash_init(
        armor            => 1,
        meta_interactive => 0,
        batch            => 1,
    );

    # set the passphrase so we don't need to type
    $gnupg->passphrase( $self->gpg_pass_phrase );

    return $gnupg;
}

# Setup the communicational channel for gpg interface
sub _gpg_channels {
    my $self   = shift;
    my %params = @_;

    my %ghandles;

    for my $p (qw{ stdin stdout stderr }) {
        $ghandles{$p} =
          $params{$p} ? IO::Handle->new( $params{$p} ) : IO::Handle->new();
    }

    # set in/out/err
    my $handles = GnuPG::Handles->new(%ghandles);

    return $handles;
}

# Encrypt a message
sub clearsign_msg {
    my $self = shift;
    my @msgs = @_;

    # setup gpg interface: stdin, stdout and stderr
    my $handles = $self->_gpg_channels;
    my $in      = $handles->stdin;
    my $out     = $handles->stdout;
    my $err     = $handles->stderr;
    my $pid     = $self->_gpg->wrap_call(
        commands => [ qw{ --clearsign --local-user }, $self->user ],
        handles  => $handles,
    );

    # send message to be crypted to the input
    print $in @msgs;
    close $in;

    # read crypted msg from the output
    my $crypted = do { local $/; <$out> };
    close $out;

    # read errors
    my $error = do { local $/; <$err> };
    close $err;
    croak "An error occured: $error" if $error;

    # return crypted msg
    waitpid $pid, 0;
    return $crypted;
}

1;

__END__

=pod

=encoding UTF-8

=head1 NAME

Gorjun - Perl Client to L<gorjun daemon|https://github.com/subutai-io/cdn>

=head1 VERSION

Version 0.1

=head1 INSTALLATION

On debian the easy way is:

 # install cpanm an local::lib
 sudo apt-get install cpanminus liblocal-lib-perl
 # clone repo
 git clone https://github.com/marcoarthur/gorjun-perl-cli.git
 # install dependencies
 cd gorjun-perl-cli
 cpanm --installdeps .

For any distro just read about installing 
L<cpanm|https://github.com/miyagawa/cpanminus> 
and L<local lib|https://metacpan.org/pod/local::lib>
then run:

 git clone https://github.com/marcoarthur/gorjun-perl-cli.git
 # install dependencies
 cd gorjun-perl-cli
 cpanm --installdeps .

=head1 SYNOPSIS

 my $g = Gorjun->new(
    host => 'http://127.0.0.1',
    port => '8080',
    user => 'Tester_User',
    gpg_pass_phrase => 'secret', # needed passphrase for gpg signing
 );

 # Get token
 $g->get_token( message => 'Some message', user => 'Tester_User' );

=head1 DESCRIPTION

This modules operates gorjun daemon using its rest interface
so you can control gorjun actions in a Perl environment, sending, receiving
messages. Useful for integration tests as well as system administration of
gorjun. There are high level operations, as well as low level.

=head2 High Level

Those are methods that performs a grouped set of low level operations. Such as
register a new user, upload a file, or sign an uploaded file. All of these won't
be executed with only one REST call, and will need to send several queries in
order to complete. All high level operations and their arguments are listed
below:

=over 4

=item *

register( name => $n, key => $k )

=item *

get_authid( user => $u )

=item *

get_token( user => $u, message => $m )

=item *

quota( user => $u, token => $t )

=item *

sign( signature => $s, token => $t )

=item *

upload( file => $f, token => $t , type => $t)

=back

=head2 Low Level

Those are simple one REST call, generally only completing a step of the High Level
operation. They map to the REST API exactly one-to-one:
L<https://github.com/subutai-io/cdn/wiki/v1>

And all of them are completed with C<send()> interface. Example:

 # list all raw files
 $g->send( 
        method => 'get'
        path   => '/kurjun/rest/raw/list'
 );

=head2 BUGS AND KNOWN ISSUES

Gorjun depends on gpg utility, since messages should be signed using a GPG key
L<https://www.gnupg.org/> . Not all operations can be completed without user intervention,
since not all gpg functions can be used in a scripted manner, requiring user to manually
type the pass phrase at least once.

=head2 LICENSE

This software is GNU GPL3 licensed, and can be freely copied, modified and redistributed.
